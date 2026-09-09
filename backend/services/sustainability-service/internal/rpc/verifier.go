package rpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sustainabilityv1 "github.com/carboncircuit/backend/gen/carboncircuit/sustainability/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

type ReviewManager interface {
	Queue(ctx context.Context, verifier service.Verifier, after string, limit int) (service.ClaimPage, error)
	Review(ctx context.Context, verifier service.Verifier, claimID uuid.UUID) (service.ReviewView, error)
	Decide(ctx context.Context, verifier service.Verifier, claimID uuid.UUID, decision service.Decision) (service.ReviewView, error)
}

var outcomeNames = map[sustainabilityv1.DecisionOutcome]domain.DecisionOutcome{
	sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_APPROVED:                   domain.DecisionApproved,
	sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_REJECTED:                   domain.DecisionRejected,
	sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_MORE_INFORMATION_REQUESTED: domain.DecisionMoreInformationRequested,
}

var outcomeCodes = map[domain.DecisionOutcome]sustainabilityv1.DecisionOutcome{
	domain.DecisionApproved:                 sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_APPROVED,
	domain.DecisionRejected:                 sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_REJECTED,
	domain.DecisionMoreInformationRequested: sustainabilityv1.DecisionOutcome_DECISION_OUTCOME_MORE_INFORMATION_REQUESTED,
}

func (s *SustainabilityServer) verifier(ctx context.Context) (service.Verifier, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present {
		return service.Verifier{}, status.Error(codes.Unauthenticated, "a verified caller is required")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return service.Verifier{}, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	return service.Verifier{
		UserID:       userID,
		Name:         verified.Name,
		PlatformRole: verified.PlatformRole,
	}, nil
}

func (s *SustainabilityServer) ReviewQueue(
	ctx context.Context,
	request *sustainabilityv1.ReviewQueueRequest,
) (*sustainabilityv1.ReviewQueueResponse, error) {
	who, err := s.verifier(ctx)
	if err != nil {
		return nil, err
	}

	page, err := s.reviews.Queue(ctx, who, request.GetAfter(), int(request.GetLimit()))
	if err != nil {
		return nil, translate(err)
	}

	claims := make([]*sustainabilityv1.Claim, 0, len(page.Claims))
	for _, claim := range page.Claims {
		claims = append(claims, claimMessage(service.ClaimView{Claim: claim}))
	}

	return &sustainabilityv1.ReviewQueueResponse{
		Claims:     claims,
		NextCursor: page.NextCursor,
	}, nil
}

func (s *SustainabilityServer) ReviewClaim(
	ctx context.Context,
	request *sustainabilityv1.ReviewClaimRequest,
) (*sustainabilityv1.ReviewClaimResponse, error) {
	who, err := s.verifier(ctx)
	if err != nil {
		return nil, err
	}

	claimID, err := uuid.Parse(request.GetClaimId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "claim id is not a valid identifier")
	}

	view, err := s.reviews.Review(ctx, who, claimID)
	if err != nil {
		return nil, translate(err)
	}

	return &sustainabilityv1.ReviewClaimResponse{
		Claim:     claimMessage(service.ClaimView{Claim: view.Claim, Evidence: view.Evidence}),
		AiReview:  aiReviewMessage(view),
		Decisions: decisionMessages(view.Decisions),
	}, nil
}

func (s *SustainabilityServer) DecideClaim(
	ctx context.Context,
	request *sustainabilityv1.DecideClaimRequest,
) (*sustainabilityv1.DecideClaimResponse, error) {
	who, err := s.verifier(ctx)
	if err != nil {
		return nil, err
	}

	claimID, err := uuid.Parse(request.GetClaimId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "claim id is not a valid identifier")
	}

	outcome, known := outcomeNames[request.GetOutcome()]
	if !known {
		return nil, status.Error(codes.InvalidArgument, "outcome must be a known decision")
	}

	if request.GetIdempotencyKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	view, err := s.reviews.Decide(ctx, who, claimID, service.Decision{
		Outcome:        outcome,
		ApprovedAmount: request.GetApprovedAmount(),
		Reason:         request.GetReason(),
		IdempotencyKey: request.GetIdempotencyKey(),
		RequestBody:    canonicalDecision(request),
	})
	if err != nil {
		return nil, translate(err)
	}

	return &sustainabilityv1.DecideClaimResponse{
		Claim:     claimMessage(service.ClaimView{Claim: view.Claim}),
		Decisions: decisionMessages(view.Decisions),
	}, nil
}

func canonicalDecision(request *sustainabilityv1.DecideClaimRequest) []byte {
	return []byte(request.GetClaimId() + "\x1f" +
		request.GetOutcome().String() + "\x1f" +
		request.GetApprovedAmount() + "\x1f" +
		request.GetReason())
}

func aiReviewMessage(view service.ReviewView) *sustainabilityv1.AIReview {
	if view.AIReview == nil {
		return nil
	}

	review := view.AIReview

	confidence := ""
	if review.Confidence != nil {
		confidence = *review.Confidence
	}

	return &sustainabilityv1.AIReview{
		Assessment: string(review.Assessment),
		Confidence: confidence,
		Narrative:  review.Narrative,
		AssessedBy: review.AssessedBy,
		AssessedAt: review.AssessedAt.UTC().Format(time.RFC3339),
	}
}

func decisionMessages(decisions []domain.ClaimDecision) []*sustainabilityv1.ClaimDecision {
	messages := make([]*sustainabilityv1.ClaimDecision, 0, len(decisions))

	for _, decision := range decisions {
		amount := ""
		if decision.ApprovedAmount != nil {
			amount = *decision.ApprovedAmount
		}

		messages = append(messages, &sustainabilityv1.ClaimDecision{
			Id:             decision.ID.String(),
			VerifierUserId: decision.VerifierUserID.String(),
			VerifierName:   decision.VerifierName,
			Outcome:        outcomeCodes[decision.Outcome],
			ApprovedAmount: amount,
			Reason:         decision.Reason,
			DecidedAt:      decision.DecidedAt.UTC().Format(time.RFC3339),
		})
	}

	return messages
}
