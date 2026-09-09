package rpc

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	sustainabilityv1 "github.com/carboncircuit/backend/gen/carboncircuit/sustainability/v1"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/ceiling"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

type ClaimManager interface {
	Submit(ctx context.Context, actor service.Actor, submission service.Submission) (service.ClaimView, error)
	Get(ctx context.Context, actor service.Actor, claimID uuid.UUID) (service.ClaimView, error)
	List(ctx context.Context, actor service.Actor, status domain.ClaimStatus, after string, limit int) (service.ClaimPage, error)
	PreviewCeiling(ctx context.Context, actor service.Actor, submission service.Submission) (service.CeilingPreview, error)
}

type SustainabilityServer struct {
	sustainabilityv1.UnimplementedSustainabilityServiceServer

	database *gorm.DB
	claims   ClaimManager
	reviews  ReviewManager
	logger   *slog.Logger
	revision string
}

func NewSustainabilityServer(
	database *gorm.DB,
	claims ClaimManager,
	reviews ReviewManager,
	logger *slog.Logger,
	revision string,
) *SustainabilityServer {
	return &SustainabilityServer{
		database: database,
		claims:   claims,
		reviews:  reviews,
		logger:   logger,
		revision: revision,
	}
}

func (s *SustainabilityServer) Ping(
	ctx context.Context,
	_ *sustainabilityv1.PingRequest,
) (*sustainabilityv1.PingResponse, error) {
	reachable := s.databaseReachable(ctx)

	var factors int64
	if reachable {
		s.database.Table("sustainability.reference_factors").
			Where("deleted_at IS NULL").Count(&factors)
	}

	return &sustainabilityv1.PingResponse{
		Service:              "sustainability-service",
		Revision:             s.revision,
		DatabaseReachable:    reachable,
		ReferenceFactorCount: int32(factors),
	}, nil
}

func (s *SustainabilityServer) databaseReachable(ctx context.Context) bool {
	pool, err := s.database.DB()
	if err != nil {
		return false
	}
	return pool.PingContext(ctx) == nil
}

func (s *SustainabilityServer) SubmitClaim(
	ctx context.Context,
	request *sustainabilityv1.SubmitClaimRequest,
) (*sustainabilityv1.SubmitClaimResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	submission, err := submissionFrom(request)
	if err != nil {
		return nil, err
	}

	submission.RequestBody = canonicalClaimRequest(request)

	view, err := s.claims.Submit(ctx, actor, submission)
	if err != nil {
		return nil, translate(err)
	}

	return &sustainabilityv1.SubmitClaimResponse{
		Claim:            claimMessage(view),
		AlreadySubmitted: view.Replayed,
	}, nil
}

func submissionFrom(request *sustainabilityv1.SubmitClaimRequest) (service.Submission, error) {
	if request.GetIdempotencyKey() == "" {
		return service.Submission{}, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	facilityID, err := uuid.Parse(request.GetFacilityId())
	if err != nil {
		return service.Submission{}, status.Error(codes.InvalidArgument, "facility id is not a valid identifier")
	}

	activity, known := activityFrom(request.GetActivityType())
	if !known {
		return service.Submission{}, status.Error(codes.InvalidArgument, "activity type must be a known activity")
	}

	periodStart, err := time.Parse(time.DateOnly, request.GetPeriodStart())
	if err != nil {
		return service.Submission{}, status.Error(codes.InvalidArgument, "period start must be a calendar date")
	}

	periodEnd, err := time.Parse(time.DateOnly, request.GetPeriodEnd())
	if err != nil {
		return service.Submission{}, status.Error(codes.InvalidArgument, "period end must be a calendar date")
	}

	evidenceIDs := make([]uuid.UUID, 0, len(request.GetEvidenceIds()))
	for _, candidate := range request.GetEvidenceIds() {
		parsed, parseErr := uuid.Parse(candidate)
		if parseErr != nil {
			return service.Submission{}, status.Errorf(codes.InvalidArgument,
				"evidence id %q is not a valid identifier", candidate)
		}
		evidenceIDs = append(evidenceIDs, parsed)
	}

	return service.Submission{
		FacilityID:      facilityID,
		ActivityType:    activity,
		VintageYear:     int(request.GetVintageYear()),
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		DeclaredFigures: request.GetDeclaredFigures(),
		RequestedAmount: request.GetRequestedAmount(),
		EvidenceIDs:     evidenceIDs,
		Attested:        request.GetExclusivityAttested(),
		IdempotencyKey:  request.GetIdempotencyKey(),
	}, nil
}

func canonicalClaimRequest(request *sustainabilityv1.SubmitClaimRequest) []byte {
	figures := make([]string, 0, len(request.GetDeclaredFigures()))
	for name, value := range request.GetDeclaredFigures() {
		figures = append(figures, name+"="+strings.TrimSpace(value))
	}
	sort.Strings(figures)

	evidence := append([]string(nil), request.GetEvidenceIds()...)
	sort.Strings(evidence)

	parts := []string{
		request.GetFacilityId(),
		request.GetActivityType().String(),
		strconv.Itoa(int(request.GetVintageYear())),
		request.GetPeriodStart(),
		request.GetPeriodEnd(),
		strings.Join(figures, ","),
		strings.TrimSpace(request.GetRequestedAmount()),
		strings.Join(evidence, ","),
		strconv.FormatBool(request.GetExclusivityAttested()),
	}

	return []byte(strings.Join(parts, "\x1f"))
}

func (s *SustainabilityServer) ListClaims(
	ctx context.Context,
	request *sustainabilityv1.ListClaimsRequest,
) (*sustainabilityv1.ListClaimsResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.claims.List(
		ctx, actor, statusFrom(request.GetStatus()), request.GetAfter(), int(request.GetLimit()),
	)
	if err != nil {
		return nil, translate(err)
	}

	claims := make([]*sustainabilityv1.Claim, 0, len(page.Claims))
	for _, claim := range page.Claims {
		claims = append(claims, claimMessage(service.ClaimView{Claim: claim}))
	}

	return &sustainabilityv1.ListClaimsResponse{
		Claims:     claims,
		NextCursor: page.NextCursor,
	}, nil
}

func (s *SustainabilityServer) GetClaim(
	ctx context.Context,
	request *sustainabilityv1.GetClaimRequest,
) (*sustainabilityv1.GetClaimResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	claimID, err := uuid.Parse(request.GetClaimId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "claim id is not a valid identifier")
	}

	view, err := s.claims.Get(ctx, actor, claimID)
	if err != nil {
		return nil, translate(err)
	}

	return &sustainabilityv1.GetClaimResponse{Claim: claimMessage(view)}, nil
}

func (s *SustainabilityServer) PreviewCeiling(
	ctx context.Context,
	request *sustainabilityv1.PreviewCeilingRequest,
) (*sustainabilityv1.PreviewCeilingResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	facilityID, err := uuid.Parse(request.GetFacilityId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "facility id is not a valid identifier")
	}

	activity, known := activityFrom(request.GetActivityType())
	if !known {
		return nil, status.Error(codes.InvalidArgument, "activity type must be a known activity")
	}

	periodStart, err := time.Parse(time.DateOnly, request.GetPeriodStart())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "period start must be a calendar date")
	}

	periodEnd, err := time.Parse(time.DateOnly, request.GetPeriodEnd())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "period end must be a calendar date")
	}

	preview, err := s.claims.PreviewCeiling(ctx, actor, service.Submission{
		FacilityID:   facilityID,
		ActivityType: activity,
		VintageYear:  int(request.GetVintageYear()),
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	})
	if err != nil {
		return nil, translate(err)
	}

	return &sustainabilityv1.PreviewCeilingResponse{
		Ceiling:        preview.Ceiling,
		VintageCeiling: preview.VintageCeiling,
		Consumed:       preview.Consumed,
		Remaining:      preview.Remaining,
		PeriodCeiling:  preview.PeriodCeiling,
		CapacityBasis:  preview.CapacityBasis,
		CapacitySource: preview.CapacitySource,
		DiscountFactor: preview.DiscountFactor,
		ReferenceValue: preview.ReferenceValue,
		GridRegion:     preview.GridRegion,
		PeriodDays:     int32(preview.PeriodDays),
		VintageDays:    int32(preview.VintageDays),
	}, nil
}

func (s *SustainabilityServer) actor(ctx context.Context) (service.Actor, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return service.Actor{}, status.Error(codes.Unauthenticated, "a verified organization is required")
	}
	return actorFrom(verified)
}

func actorFrom(verified servicetoken.Caller) (service.Actor, error) {
	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	return service.Actor{
		OrganizationID:    organizationID,
		UserID:            userID,
		OrganizationType:  verified.OrganizationType,
		OrganizationState: verified.OrganizationState,
		PlanTier:          verified.PlanTier,
	}, nil
}

var refusalReasons = []struct {
	sentinel error
	reason   string
}{
	{service.ErrActivityUnsupported, "ACTIVITY_UNSUPPORTED"},
	{service.ErrEvidenceRequired, "EVIDENCE_REQUIRED"},
	{service.ErrTooMuchEvidence, "TOO_MUCH_EVIDENCE"},
	{service.ErrEvidenceUnusable, "EVIDENCE_UNUSABLE"},
	{service.ErrAttestationRequired, "ATTESTATION_REQUIRED"},
	{service.ErrVintageCapacityLeft, "VINTAGE_CAPACITY_EXHAUSTED"},
	{service.ErrClaimNotInReview, "NOT_IN_REVIEW"},
	{service.ErrAboveCeiling, "ABOVE_CEILING"},
	{service.ErrAboveRequested, "ABOVE_REQUESTED"},
	{service.ErrReasonTooShort, "REASON_TOO_SHORT"},
	{service.ErrAlreadyDecided, "ALREADY_DECIDED"},
	{service.ErrApprovalNotPositive, "APPROVAL_NOT_POSITIVE"},
	{ceiling.ErrVintageExhausted, "VINTAGE_CAPACITY_EXHAUSTED"},
	{ceiling.ErrCapacityUnknown, "CAPACITY_UNKNOWN"},
	{ceiling.ErrPeriodOutsideVintage, "PERIOD_OUTSIDE_VINTAGE"},
	{ceiling.ErrPeriodEmpty, "PERIOD_EMPTY"},
	{ceiling.ErrFactorUnknown, "FACTOR_UNKNOWN"},
	{ceiling.ErrDiscountUnknown, "DISCOUNT_UNKNOWN"},
}

func refused(err error, reason string) error {
	reported := status.New(codes.FailedPrecondition, err.Error())

	detailed, attachErr := reported.WithDetails(&errdetails.ErrorInfo{
		Reason: reason,
		Domain: events.ClaimRefusalDomain,
	})
	if attachErr != nil {
		return reported.Err()
	}

	return detailed.Err()
}

func translate(err error) error {
	for _, candidate := range refusalReasons {
		if errors.Is(err, candidate.sentinel) {
			return refused(err, candidate.reason)
		}
	}

	switch {
	case errors.Is(err, service.ErrClaimNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrFacilityUnknown):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrOrganizationReadOnly),
		errors.Is(err, service.ErrNotAVerifier):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrRequestInProgress):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, service.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
