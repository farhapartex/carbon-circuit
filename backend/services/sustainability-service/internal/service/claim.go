package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	submitEndpoint  = "POST /v1/claims"
	claimAggregate  = "claim"
	defaultPageSize = 25
	maximumPageSize = 100
)

var (
	ErrRequestInProgress    = errors.New("an identical request is already being processed")
	ErrIdempotencyConflict  = errors.New("idempotency key was used with a different request")
	ErrClaimNotFound        = errors.New("claim not found")
	ErrFacilityUnknown      = errors.New("facility is not registered to this organization")
	ErrOrganizationReadOnly = errors.New("this organization cannot submit claims right now")
	ErrActivityUnsupported  = errors.New("this activity type cannot yet be submitted")
	ErrEvidenceRequired     = errors.New("a claim needs at least one supporting document")
	ErrTooMuchEvidence      = errors.New("a claim may carry at most 25 supporting documents")
	ErrEvidenceUnusable     = errors.New("one or more documents did not pass scanning")
	ErrAttestationRequired  = errors.New("the exclusivity attestation is required")
	ErrCeilingExhausted     = errors.New("this facility's ceiling for that vintage is already used")
)

type Actor struct {
	OrganizationID    uuid.UUID
	UserID            uuid.UUID
	OrganizationType  string
	OrganizationState string
	PlanTier          string
}

func (a Actor) maySubmit() error {
	if a.OrganizationState == "read_only" || a.OrganizationState == "suspended" {
		return ErrOrganizationReadOnly
	}
	if a.OrganizationType == "credit_buyer" {
		return ErrOrganizationReadOnly
	}
	return nil
}

type Facility struct {
	ID                    uuid.UUID
	Name                  string
	GridRegion            string
	CeilingDiscountFactor string
	DeclaredEnergyKwh     string
	AttestedEnergyKwh     string
}

func (f Facility) capacity() (ceiling.Capacity, error) {
	if f.AttestedEnergyKwh != "" {
		attested, err := decimal.NewFromString(f.AttestedEnergyKwh)
		if err == nil && attested.GreaterThan(decimal.Zero) {
			return ceiling.Capacity{Value: attested, Source: ceiling.Attested}, nil
		}
	}

	declared, err := decimal.NewFromString(f.DeclaredEnergyKwh)
	if err != nil || declared.LessThanOrEqual(decimal.Zero) {
		return ceiling.Capacity{}, ceiling.ErrCapacityUnknown
	}

	return ceiling.Capacity{Value: declared, Source: ceiling.Declared}, nil
}

type FacilityResolver interface {
	Facility(ctx context.Context, facilityID uuid.UUID) (Facility, error)
}

type ResolvedEvidence struct {
	DocumentID  uuid.UUID
	Usable      bool
	Refusal     string
	ContentHash string
	FileName    string
	MediaType   string
	ByteSize    int64
	PageCount   *int
}

type EvidenceResolver interface {
	Resolve(ctx context.Context, documentIDs []uuid.UUID) ([]ResolvedEvidence, error)
}

type Submission struct {
	FacilityID      uuid.UUID
	ActivityType    domain.ActivityType
	VintageYear     int
	PeriodStart     time.Time
	PeriodEnd       time.Time
	DeclaredFigures map[string]string
	RequestedAmount string
	EvidenceIDs     []uuid.UUID
	Attested        bool
	IdempotencyKey  string
	RequestBody     []byte
}

func (s Submission) validate() error {
	if s.ActivityType != domain.RenewableEnergy {
		return fmt.Errorf("%w: %s has no recorded facility capacity to cap it",
			ErrActivityUnsupported, s.ActivityType)
	}
	if !s.Attested {
		return ErrAttestationRequired
	}
	if len(s.EvidenceIDs) == 0 {
		return ErrEvidenceRequired
	}
	if len(s.EvidenceIDs) > domain.MaximumEvidenceCount {
		return ErrTooMuchEvidence
	}

	requested, err := decimal.NewFromString(s.RequestedAmount)
	if err != nil || requested.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("requested amount must be a positive decimal")
	}

	return nil
}

type ClaimView struct {
	Claim    domain.Claim
	Evidence []domain.ClaimEvidence
	Replayed bool
}

type ClaimService struct {
	database   *gorm.DB
	claims     repository.ClaimStore
	references repository.ReferenceStore
	facilities FacilityResolver
	evidence   EvidenceResolver
	logger     *slog.Logger
}

func NewClaimService(
	handle *gorm.DB,
	claims repository.ClaimStore,
	references repository.ReferenceStore,
	facilities FacilityResolver,
	evidence EvidenceResolver,
	logger *slog.Logger,
) *ClaimService {
	return &ClaimService{
		database:   handle,
		claims:     claims,
		references: references,
		facilities: facilities,
		evidence:   evidence,
		logger:     logger,
	}
}

func (s *ClaimService) Submit(
	ctx context.Context,
	actor Actor,
	submission Submission,
) (ClaimView, error) {
	if err := actor.maySubmit(); err != nil {
		return ClaimView{}, err
	}
	if err := submission.validate(); err != nil {
		return ClaimView{}, err
	}

	facility, err := s.facilities.Facility(ctx, submission.FacilityID)
	if err != nil {
		return ClaimView{}, err
	}

	attachments, err := s.confirmEvidence(ctx, submission.EvidenceIDs)
	if err != nil {
		return ClaimView{}, err
	}

	claimID, err := uuid.NewV7()
	if err != nil {
		return ClaimView{}, fmt.Errorf("generate claim id: %w", err)
	}

	var view ClaimView

	err = database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		computed, err := s.computeCeiling(tx, facility, submission)
		if err != nil {
			return err
		}

		created, replayed, err := s.persist(
			tx, actor, claimID, facility, submission, computed, attachments,
		)
		if err != nil {
			return err
		}

		view = created
		view.Replayed = replayed
		return nil
	})
	if err != nil {
		return ClaimView{}, err
	}

	return view, nil
}

type computedCeiling struct {
	Result   ceiling.Result
	Capacity ceiling.Capacity
	Factor   domain.ReferenceFactor
	Discount decimal.Decimal
}

func (s *ClaimService) computeCeiling(
	tx database.Tx,
	facility Facility,
	submission Submission,
) (computedCeiling, error) {
	capacity, err := facility.capacity()
	if err != nil {
		return computedCeiling{}, err
	}

	factor, found, err := s.references.Effective(
		tx, domain.GridEmission, facility.GridRegion, submission.PeriodStart,
	)
	if err != nil {
		return computedCeiling{}, err
	}
	if !found {
		return computedCeiling{}, fmt.Errorf("%w: no grid factor for %s in %d",
			ceiling.ErrFactorUnknown, facility.GridRegion, submission.VintageYear)
	}

	factorValue, err := decimal.NewFromString(factor.Factor)
	if err != nil {
		return computedCeiling{}, fmt.Errorf("reference factor %s is unusable", factor.ID)
	}

	discount, err := decimal.NewFromString(facility.CeilingDiscountFactor)
	if err != nil {
		return computedCeiling{}, ceiling.ErrDiscountUnknown
	}

	result, err := ceiling.ForRenewableEnergy(ceiling.Inputs{
		VintageYear:    submission.VintageYear,
		Period:         ceiling.Period{Start: submission.PeriodStart, End: submission.PeriodEnd},
		Capacity:       capacity,
		ReferenceValue: factorValue,
		DiscountFactor: discount,
	})
	if err != nil {
		return computedCeiling{}, err
	}

	return computedCeiling{
		Result:   result,
		Capacity: capacity,
		Factor:   factor,
		Discount: discount,
	}, nil
}

func (s *ClaimService) confirmEvidence(
	ctx context.Context,
	documentIDs []uuid.UUID,
) ([]ResolvedEvidence, error) {
	resolved, err := s.evidence.Resolve(ctx, documentIDs)
	if err != nil {
		return nil, err
	}

	if len(resolved) != len(documentIDs) {
		return nil, fmt.Errorf("%w: evidence answered for %d of %d documents",
			ErrEvidenceUnusable, len(resolved), len(documentIDs))
	}

	for _, document := range resolved {
		if !document.Usable {
			return nil, fmt.Errorf("%w: %s", ErrEvidenceUnusable, document.Refusal)
		}
	}

	return resolved, nil
}

func (s *ClaimService) persist(
	tx database.Tx,
	actor Actor,
	claimID uuid.UUID,
	facility Facility,
	submission Submission,
	computed computedCeiling,
	attachments []ResolvedEvidence,
) (ClaimView, bool, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForOrganization(actor.OrganizationID),
		Endpoint: submitEndpoint,
		Key:      submission.IdempotencyKey,
		Body:     submission.RequestBody,
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return ClaimView{}, false, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return ClaimView{}, false, ErrIdempotencyConflict
	case err != nil:
		return ClaimView{}, false, err
	}

	if reservation.IsReplay() {
		var replayed ClaimView
		if err := json.Unmarshal(reservation.Replay.Body, &replayed); err != nil {
			return ClaimView{}, false, fmt.Errorf("decode replayed claim: %w", err)
		}
		return replayed, true, nil
	}

	declared := map[string]string{}
	for name, value := range submission.DeclaredFigures {
		declared[name] = value
	}
	declared["grid_region"] = facility.GridRegion

	figures, err := json.Marshal(declared)
	if err != nil {
		return ClaimView{}, false, fmt.Errorf("encode declared figures: %w", err)
	}

	requested, _ := decimal.NewFromString(submission.RequestedAmount)
	now := time.Now().UTC()

	claim := domain.Claim{
		OrganizationID:        actor.OrganizationID,
		SubmittedByUserID:     actor.UserID,
		FacilityID:            facility.ID,
		FacilityName:          facility.Name,
		ActivityType:          submission.ActivityType,
		VintageYear:           submission.VintageYear,
		PeriodStart:           submission.PeriodStart,
		PeriodEnd:             submission.PeriodEnd,
		DeclaredFigures:       database.JSONDocument(figures),
		RequestedAmount:       requested.StringFixed(ceiling.Places),
		ComputedCeiling:       computed.Result.Ceiling.StringFixed(ceiling.Places),
		CapacityBasis:         computed.Capacity.Value.StringFixed(ceiling.Places),
		CapacitySource:        computed.Capacity.Source,
		DiscountFactor:        computed.Discount.StringFixed(2),
		ReferenceFactorID:     computed.Factor.ID,
		ReferenceFactorValue:  computed.Factor.Factor,
		ReferenceLookupKey:    computed.Factor.LookupKey,
		Status:                domain.Submitted,
		Priority:              PriorityFor(requested, computed.Result.Ceiling, facility.CeilingDiscountFactor),
		RequiresDualApproval:  RequiresDualApproval(computed.Result.Ceiling),
		ExclusivityAttestedAt: now,
		ExclusivityAttestedBy: actor.UserID,
	}
	claim.ID = claimID

	if err := s.claims.Insert(tx, &claim); err != nil {
		return ClaimView{}, false, err
	}

	linked := make([]domain.ClaimEvidence, 0, len(attachments))
	for _, attachment := range attachments {
		linked = append(linked, domain.ClaimEvidence{
			OrganizationID: actor.OrganizationID,
			ClaimID:        claimID,
			EvidenceID:     attachment.DocumentID,
			ContentHash:    attachment.ContentHash,
			FileName:       attachment.FileName,
			MediaType:      attachment.MediaType,
			ByteSize:       attachment.ByteSize,
			PageCount:      attachment.PageCount,
		})
	}

	if err := s.claims.AttachEvidence(tx, linked); err != nil {
		return ClaimView{}, false, err
	}

	hashes := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		hashes = append(hashes, attachment.ContentHash)
	}

	if _, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: claimAggregate,
		AggregateID:   claimID,
		EventType:     events.TopicClaimSubmitted,
		Payload: events.ClaimSubmitted{
			ClaimID:              claimID.String(),
			OrganizationID:       actor.OrganizationID.String(),
			FacilityID:           facility.ID.String(),
			ActivityType:         string(submission.ActivityType),
			VintageYear:          submission.VintageYear,
			PeriodStart:          submission.PeriodStart.Format(time.DateOnly),
			PeriodEnd:            submission.PeriodEnd.Format(time.DateOnly),
			RequestedAmount:      claim.RequestedAmount,
			ComputedCeiling:      claim.ComputedCeiling,
			Priority:             string(claim.Priority),
			RequiresDualApproval: claim.RequiresDualApproval,
			EvidenceHashes:       hashes,
			SubmittedAt:          now.Format(time.RFC3339),
		},
	}); err != nil {
		return ClaimView{}, false, err
	}

	view := ClaimView{Claim: claim, Evidence: linked}

	body, err := json.Marshal(view)
	if err != nil {
		return ClaimView{}, false, fmt.Errorf("encode idempotent response: %w", err)
	}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     201,
		Body:       body,
		ResourceID: &claimID,
	}); err != nil {
		return ClaimView{}, false, err
	}

	return view, false, nil
}

func tenancy(actor Actor) database.TenantContext {
	return database.TenantContext{
		UserID:         actor.UserID.String(),
		OrganizationID: actor.OrganizationID.String(),
	}
}
