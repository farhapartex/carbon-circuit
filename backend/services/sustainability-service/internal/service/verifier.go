package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/ceiling"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/repository"
)

const (
	decideEndpoint = "POST /v1/verifier/claims/{id}/decision"
	VerifierRole   = "verifier"
)

var (
	ErrNotAVerifier        = errors.New("this action requires the verifier role")
	ErrClaimNotInReview    = errors.New("this claim is not awaiting a verification decision")
	ErrAboveCeiling        = errors.New("a claim cannot be approved above its computed ceiling")
	ErrAboveRequested      = errors.New("a claim cannot be approved above the amount requested")
	ErrReasonTooShort      = errors.New("a rejection or information request needs a reason of at least 40 characters")
	ErrAlreadyDecided      = errors.New("this verifier has already decided this claim")
	ErrSecondOpinionOwn    = errors.New("a second approval must come from a different verifier")
	ErrApprovalNotPositive = errors.New("an approved amount must be greater than zero")
)

type Verifier struct {
	UserID       uuid.UUID
	Name         string
	PlatformRole string
}

func (v Verifier) mayReview() error {
	if v.PlatformRole != VerifierRole {
		return ErrNotAVerifier
	}
	if v.UserID == uuid.Nil {
		return ErrNotAVerifier
	}
	return nil
}

func (v Verifier) tenancy() database.TenantContext {
	return database.TenantContext{
		UserID:       v.UserID.String(),
		PlatformRole: v.PlatformRole,
	}
}

type Decision struct {
	Outcome        domain.DecisionOutcome
	ApprovedAmount string
	Reason         string
	IdempotencyKey string
	RequestBody    []byte
}

func (d Decision) validate(claim domain.Claim) error {
	switch d.Outcome {
	case domain.DecisionApproved:
		approved, err := decimal.NewFromString(d.ApprovedAmount)
		if err != nil || approved.LessThanOrEqual(decimal.Zero) {
			return ErrApprovalNotPositive
		}

		ceilingAmount, err := decimal.NewFromString(claim.ComputedCeiling)
		if err != nil {
			return fmt.Errorf("claim ceiling is unusable")
		}
		if approved.GreaterThan(ceilingAmount) {
			return fmt.Errorf("%w: %s exceeds %s",
				ErrAboveCeiling, approved.StringFixed(ceiling.Places), claim.ComputedCeiling)
		}

		requested, err := decimal.NewFromString(claim.RequestedAmount)
		if err != nil {
			return fmt.Errorf("claim requested amount is unusable")
		}
		if approved.GreaterThan(requested) {
			return fmt.Errorf("%w: %s exceeds %s",
				ErrAboveRequested, approved.StringFixed(ceiling.Places), claim.RequestedAmount)
		}

		return nil

	case domain.DecisionRejected, domain.DecisionMoreInformationRequested:
		if len(strings.TrimSpace(d.Reason)) < domain.MinimumReasonLength {
			return ErrReasonTooShort
		}
		return nil

	default:
		return fmt.Errorf("decision outcome %q is not a known outcome", d.Outcome)
	}
}

type ReviewView struct {
	Claim     domain.Claim
	Evidence  []domain.ClaimEvidence
	AIReview  *domain.ClaimAIReview
	Decisions []domain.ClaimDecision
}

type VerifierService struct {
	database *gorm.DB
	claims   repository.ClaimStore
	logger   *slog.Logger
}

func NewVerifierService(
	handle *gorm.DB,
	claims repository.ClaimStore,
	logger *slog.Logger,
) *VerifierService {
	return &VerifierService{database: handle, claims: claims, logger: logger}
}

func (s *VerifierService) Queue(
	ctx context.Context,
	verifier Verifier,
	after string,
	limit int,
) (ClaimPage, error) {
	if err := verifier.mayReview(); err != nil {
		return ClaimPage{}, err
	}

	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maximumPageSize {
		limit = maximumPageSize
	}

	var page ClaimPage

	err := database.WithinTenant(ctx, s.database, verifier.tenancy(), func(tx database.Tx) error {
		claims, err := s.claims.Queue(tx, after, limit+1)
		if err != nil {
			return err
		}

		if len(claims) > limit {
			page.NextCursor = claims[limit-1].ID.String()
			claims = claims[:limit]
		}

		page.Claims = claims
		return nil
	})
	if err != nil {
		return ClaimPage{}, err
	}

	return page, nil
}

func (s *VerifierService) Review(
	ctx context.Context,
	verifier Verifier,
	claimID uuid.UUID,
) (ReviewView, error) {
	if err := verifier.mayReview(); err != nil {
		return ReviewView{}, err
	}

	var view ReviewView

	err := database.WithinTenant(ctx, s.database, verifier.tenancy(), func(tx database.Tx) error {
		claim, found, err := s.claims.FindForReview(tx, claimID)
		if err != nil {
			return err
		}
		if !found {
			return ErrClaimNotFound
		}

		evidence, err := s.claims.Evidence(tx, claim.OrganizationID, claimID)
		if err != nil {
			return err
		}

		decisions, err := s.claims.Decisions(tx, claimID)
		if err != nil {
			return err
		}

		review, present, err := s.claims.AIReview(tx, claim.OrganizationID, claimID)
		if err != nil {
			return err
		}

		view = ReviewView{Claim: claim, Evidence: evidence, Decisions: decisions}
		if present {
			view.AIReview = &review
		}

		return nil
	})
	if err != nil {
		return ReviewView{}, err
	}

	return view, nil
}

func (s *VerifierService) Decide(
	ctx context.Context,
	verifier Verifier,
	claimID uuid.UUID,
	decision Decision,
) (ReviewView, error) {
	if err := verifier.mayReview(); err != nil {
		return ReviewView{}, err
	}

	var view ReviewView

	err := database.WithinTenant(ctx, s.database, verifier.tenancy(), func(tx database.Tx) error {
		claim, found, err := s.claims.FindForReview(tx, claimID)
		if err != nil {
			return err
		}
		if !found {
			return ErrClaimNotFound
		}

		if claim.Status != domain.HumanReview {
			return fmt.Errorf("%w: it is %s", ErrClaimNotInReview, claim.Status)
		}

		if err := decision.validate(claim); err != nil {
			return err
		}

		existing, err := s.claims.Decisions(tx, claimID)
		if err != nil {
			return err
		}

		for _, recorded := range existing {
			if recorded.VerifierUserID == verifier.UserID {
				return ErrAlreadyDecided
			}
		}

		recorded, err := s.record(tx, verifier, claim, decision, existing)
		if err != nil {
			return err
		}

		view = recorded
		return nil
	})
	if err != nil {
		return ReviewView{}, err
	}

	return view, nil
}

func (s *VerifierService) record(
	tx database.Tx,
	verifier Verifier,
	claim domain.Claim,
	decision Decision,
	existing []domain.ClaimDecision,
) (ReviewView, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForUser(verifier.UserID),
		Endpoint: decideEndpoint,
		Key:      decision.IdempotencyKey,
		Body:     decision.RequestBody,
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return ReviewView{}, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return ReviewView{}, ErrIdempotencyConflict
	case err != nil:
		return ReviewView{}, err
	}

	now := time.Now().UTC()

	name := verifier.Name
	if strings.TrimSpace(name) == "" {
		name = "Verifier " + verifier.UserID.String()[:8]
	}

	entry := domain.ClaimDecision{
		OrganizationID: claim.OrganizationID,
		ClaimID:        claim.ID,
		VerifierUserID: verifier.UserID,
		VerifierName:   name,
		Outcome:        decision.Outcome,
		Reason:         strings.TrimSpace(decision.Reason),
		DecidedAt:      now,
	}

	if decision.Outcome == domain.DecisionApproved {
		amount := decimal.RequireFromString(decision.ApprovedAmount).StringFixed(ceiling.Places)
		entry.ApprovedAmount = &amount
	}

	if err := s.claims.RecordDecision(tx, &entry); err != nil {
		if errors.Is(err, repository.ErrAlreadyDecided) {
			return ReviewView{}, ErrAlreadyDecided
		}
		return ReviewView{}, err
	}

	settled := append(append([]domain.ClaimDecision{}, existing...), entry)

	status, issued := outcomeOf(claim, settled)

	if err := database.AdoptTenant(tx, database.TenantContext{
		UserID:         verifier.UserID.String(),
		OrganizationID: claim.OrganizationID.String(),
		PlatformRole:   verifier.PlatformRole,
	}); err != nil {
		return ReviewView{}, err
	}

	if err := s.claims.Settle(tx, claim.ID, status, issued); err != nil {
		return ReviewView{}, err
	}

	if status != domain.HumanReview {
		if err := s.publishDecision(tx, claim, entry, status, issued); err != nil {
			return ReviewView{}, err
		}
	}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     200,
		Body:       []byte(`{"recorded":true}`),
		ResourceID: &claim.ID,
	}); err != nil {
		return ReviewView{}, err
	}

	claim.Status = status
	if issued != nil {
		claim.IssuedAmount = issued
	}

	return ReviewView{Claim: claim, Decisions: settled}, nil
}

func outcomeOf(
	claim domain.Claim,
	decisions []domain.ClaimDecision,
) (domain.ClaimStatus, *string) {
	latest := decisions[len(decisions)-1]

	switch latest.Outcome {
	case domain.DecisionRejected:
		return domain.Rejected, nil
	case domain.DecisionMoreInformationRequested:
		return domain.MoreInformationRequested, nil
	}

	approvals := make([]domain.ClaimDecision, 0, len(decisions))
	for _, decision := range decisions {
		if decision.Outcome == domain.DecisionApproved {
			approvals = append(approvals, decision)
		}
	}

	if claim.RequiresDualApproval && len(approvals) < 2 {
		return domain.HumanReview, nil
	}

	lowest := approvals[0].ApprovedAmount
	for _, approval := range approvals[1:] {
		if decimal.RequireFromString(*approval.ApprovedAmount).
			LessThan(decimal.RequireFromString(*lowest)) {
			lowest = approval.ApprovedAmount
		}
	}

	return domain.Approved, lowest
}

func (s *VerifierService) publishDecision(
	tx database.Tx,
	claim domain.Claim,
	entry domain.ClaimDecision,
	status domain.ClaimStatus,
	issued *string,
) error {
	amount := ""
	if issued != nil {
		amount = *issued
	}

	_, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: claimAggregate,
		AggregateID:   claim.ID,
		EventType:     events.TopicClaimDecisionRecorded,
		Payload: events.ClaimDecisionRecorded{
			ClaimID:        claim.ID.String(),
			OrganizationID: claim.OrganizationID.String(),
			FacilityID:     claim.FacilityID.String(),
			ActivityType:   string(claim.ActivityType),
			VintageYear:    claim.VintageYear,
			Outcome:        string(entry.Outcome),
			Status:         string(status),
			IssuedAmount:   amount,
			VerifierUserID: entry.VerifierUserID.String(),
			DecidedAt:      entry.DecidedAt.Format(time.RFC3339),
		},
	})

	return err
}

func ValidateDecision(decision Decision, claim domain.Claim) error {
	return decision.validate(claim)
}

func OutcomeOf(claim domain.Claim, decisions []domain.ClaimDecision) (domain.ClaimStatus, *string) {
	return outcomeOf(claim, decisions)
}

func MayReview(verifier Verifier) error { return verifier.mayReview() }
