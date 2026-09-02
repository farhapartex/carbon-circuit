package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-service/internal/reference"
	"github.com/carboncircuit/backend/services/provenance-service/internal/repository"
	"github.com/carboncircuit/backend/services/provenance-service/internal/score"
)

var emptyScoreComponents = database.JSONDocument("[]")

const (
	createBatchEndpoint = "POST /v1/batches"
	batchAggregate      = "batch"
	batchCreatedEvent   = "batch.created"
	defaultPageSize     = 25
	maximumPageSize     = 100
)

var (
	ErrRequestInProgress    = errors.New("an identical request is already being processed")
	ErrIdempotencyConflict  = errors.New("idempotency key was used with a different request")
	ErrFacilityUnknown      = errors.New("originating facility is not registered to this organization")
	ErrBatchNotFound        = errors.New("batch not found")
	ErrExternalIDTaken      = errors.New("external id already used by this organization")
	ErrChainTooDeep         = errors.New("component chain would exceed the maximum depth")
	ErrOrganizationReadOnly = errors.New("this organization cannot create records right now")
	ErrPlanForbidsBatches   = errors.New("this plan does not permit producing batches")
)

var (
	plainDecimal    = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)
	base62Reference = regexp.MustCompile(`^[0-9A-Za-z]{22}$`)

	productCategories = map[domain.ProductCategory]struct{}{
		domain.CategoryElectronics: {},
		domain.CategoryAgriculture: {},
		domain.CategoryPharma:      {},
		domain.CategoryTextiles:    {},
	}
)

type Actor struct {
	OrganizationID    uuid.UUID
	UserID            uuid.UUID
	OrganizationName  string
	OrganizationType  string
	PlanTier          string
	OrganizationState string
}

func (a Actor) mayProduce() error {
	if a.OrganizationState == "read_only" || a.OrganizationState == "suspended" {
		return ErrOrganizationReadOnly
	}
	if a.OrganizationType == "credit_buyer" {
		return ErrPlanForbidsBatches
	}
	return nil
}

type Facility struct {
	ID   uuid.UUID
	Name string
}

type FacilityResolver interface {
	Facility(ctx context.Context, facilityID uuid.UUID) (Facility, error)
}

type BatchDeclaration struct {
	OriginatingFacilityID uuid.UUID
	ProductCategory       domain.ProductCategory
	ComponentType         string
	LotNumber             string
	Quantity              string
	Unit                  string
	ProducedAt            time.Time
	ExternalID            string
	ParentReferences      []string
	IdempotencyKey        string
	RequestBody           []byte
}

func (d BatchDeclaration) Validate(now time.Time) error {
	if _, known := productCategories[d.ProductCategory]; !known {
		return fmt.Errorf("product category %q is not a known category", d.ProductCategory)
	}
	if strings.TrimSpace(d.ComponentType) == "" {
		return fmt.Errorf("component type is required")
	}
	if !plainDecimal.MatchString(d.Quantity) {
		return fmt.Errorf("quantity must be a decimal with up to six places")
	}
	if isZeroFigure(d.Quantity) {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if strings.TrimSpace(d.Unit) == "" {
		return fmt.Errorf("unit is required")
	}
	if d.ProducedAt.IsZero() {
		return fmt.Errorf("production date is required")
	}
	if d.ProducedAt.After(now) {
		return fmt.Errorf("a batch cannot be produced in the future")
	}
	for _, declared := range d.ParentReferences {
		if !base62Reference.MatchString(declared) {
			return fmt.Errorf("component batch reference %q is not a public batch reference", declared)
		}
	}
	return nil
}

func isZeroFigure(value string) bool {
	for _, character := range value {
		if character >= '1' && character <= '9' {
			return false
		}
	}
	return true
}

type ParentView struct {
	DeclaredReference string
	Resolved          bool
	Batch             *domain.Batch
}

type BatchView struct {
	Batch    domain.Batch
	Parents  []ParentView
	Replayed bool
}

type BatchService struct {
	database    *gorm.DB
	batches     repository.BatchStore
	checkpoints repository.CheckpointStore
	facilities  FacilityResolver
	logger      *slog.Logger
}

func NewBatchService(
	handle *gorm.DB,
	batches repository.BatchStore,
	checkpoints repository.CheckpointStore,
	facilities FacilityResolver,
	logger *slog.Logger,
) *BatchService {
	return &BatchService{
		database:    handle,
		batches:     batches,
		checkpoints: checkpoints,
		facilities:  facilities,
		logger:      logger,
	}
}

func (s *BatchService) Create(
	ctx context.Context,
	actor Actor,
	declaration BatchDeclaration,
) (BatchView, error) {
	if err := actor.mayProduce(); err != nil {
		return BatchView{}, err
	}

	now := time.Now().UTC()
	if err := declaration.Validate(now); err != nil {
		return BatchView{}, err
	}

	facility, err := s.facilities.Facility(ctx, declaration.OriginatingFacilityID)
	if err != nil {
		return BatchView{}, err
	}

	batchID, err := uuid.NewV7()
	if err != nil {
		return BatchView{}, fmt.Errorf("generate batch id: %w", err)
	}

	publicReference, err := reference.New()
	if err != nil {
		return BatchView{}, err
	}

	var view BatchView

	err = database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         actor.UserID.String(),
			OrganizationID: actor.OrganizationID.String(),
		},
		func(tx database.Tx) error {
			created, replay, workErr := s.persist(
				tx, actor, batchID, publicReference, facility, declaration,
			)
			if workErr != nil {
				return workErr
			}
			view = created
			view.Replayed = replay
			return nil
		},
	)
	if err != nil {
		return BatchView{}, err
	}

	return view, nil
}

func (s *BatchService) persist(
	tx database.Tx,
	actor Actor,
	batchID uuid.UUID,
	publicReference string,
	facility Facility,
	declaration BatchDeclaration,
) (BatchView, bool, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForOrganization(actor.OrganizationID),
		Endpoint: createBatchEndpoint,
		Key:      declaration.IdempotencyKey,
		Body:     declaration.RequestBody,
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return BatchView{}, false, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return BatchView{}, false, ErrIdempotencyConflict
	case err != nil:
		return BatchView{}, false, err
	}

	if reservation.IsReplay() {
		var replayed BatchView
		if err := json.Unmarshal(reservation.Replay.Body, &replayed); err != nil {
			return BatchView{}, false, fmt.Errorf("decode replayed batch: %w", err)
		}
		return replayed, true, nil
	}

	batch := domain.Batch{
		OrganizationID:          actor.OrganizationID,
		OriginatingFacilityID:   facility.ID,
		OriginatingFacilityName: facility.Name,
		PublicReference:         publicReference,
		ProductCategory:         declaration.ProductCategory,
		ComponentType:           strings.TrimSpace(declaration.ComponentType),
		LotNumber:               optional(declaration.LotNumber),
		Quantity:                declaration.Quantity,
		Unit:                    strings.TrimSpace(declaration.Unit),
		ProducedAt:              declaration.ProducedAt,
		ExternalID:              optional(declaration.ExternalID),
		ScoreComponents:         emptyScoreComponents,
	}
	batch.ID = batchID

	if err := s.batches.Insert(tx, &batch); err != nil {
		if errors.Is(err, repository.ErrExternalIDTaken) {
			return BatchView{}, false, ErrExternalIDTaken
		}
		return BatchView{}, false, err
	}

	parents, err := s.declareParents(tx, actor, batchID, declaration.ParentReferences)
	if err != nil {
		return BatchView{}, false, err
	}

	if err := s.rescore(tx, &batch, parents); err != nil {
		return BatchView{}, false, err
	}

	if _, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: batchAggregate,
		AggregateID:   batch.ID,
		EventType:     batchCreatedEvent,
		Payload: map[string]string{
			"batch_id":         batch.ID.String(),
			"organization_id":  batch.OrganizationID.String(),
			"public_reference": batch.PublicReference,
			"product_category": string(batch.ProductCategory),
		},
	}); err != nil {
		return BatchView{}, false, err
	}

	view := BatchView{Batch: batch, Parents: parents}

	body, err := json.Marshal(view)
	if err != nil {
		return BatchView{}, false, fmt.Errorf("encode idempotent response: %w", err)
	}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     201,
		Body:       body,
		ResourceID: &batch.ID,
	}); err != nil {
		return BatchView{}, false, err
	}

	return view, false, nil
}

func (s *BatchService) declareParents(
	tx database.Tx,
	actor Actor,
	batchID uuid.UUID,
	declared []string,
) ([]ParentView, error) {
	seen := make(map[string]bool, len(declared))
	parents := make([]ParentView, 0, len(declared))

	for _, candidate := range declared {
		if seen[candidate] {
			continue
		}
		seen[candidate] = true

		resolvedID, found, err := s.batches.ResolveReference(tx, candidate)
		if err != nil {
			return nil, err
		}

		parent := domain.BatchParent{
			OrganizationID:    actor.OrganizationID,
			BatchID:           batchID,
			DeclaredReference: candidate,
		}

		view := ParentView{DeclaredReference: candidate}

		if found {
			depth, err := s.batches.AncestorDepth(tx, resolvedID)
			if err != nil {
				return nil, err
			}
			if depth+1 >= domain.MaximumParentChainDepth {
				return nil, ErrChainTooDeep
			}

			parent.ParentBatchID = &resolvedID
			view.Resolved = true
		}

		if err := s.batches.InsertParent(tx, &parent); err != nil {
			return nil, err
		}

		if found {
			disclosable, exists, err := s.batches.Find(tx, resolvedID)
			if err != nil {
				return nil, err
			}
			if exists {
				view.Batch = &disclosable
			}
		}

		parents = append(parents, view)
	}

	return parents, nil
}

func (s *BatchService) rescore(
	tx database.Tx,
	batch *domain.Batch,
	parents []ParentView,
) error {
	checkpoints, err := s.checkpoints.ListForBatch(tx, batch.ID)
	if err != nil {
		return err
	}

	resolved := 0
	for _, parent := range parents {
		if parent.Resolved {
			resolved++
		}
	}

	computed := score.Compute(score.Input{
		Category:        batch.ProductCategory,
		Checkpoints:     scoreCheckpoints(checkpoints),
		DeclaredParents: len(parents),
		ResolvedParents: resolved,
		Now:             time.Now().UTC(),
	})

	components, err := json.Marshal(computed.Components)
	if err != nil {
		return fmt.Errorf("encode score components: %w", err)
	}

	batch.CheckpointCount = countLiving(checkpoints)
	batch.ProvenanceScore = computed.Total
	batch.ScoreComponents = components

	return s.batches.UpdateScore(tx, batch)
}

func scoreCheckpoints(checkpoints []domain.Checkpoint) []score.Checkpoint {
	mapped := make([]score.Checkpoint, 0, len(checkpoints))
	for _, checkpoint := range checkpoints {
		mapped = append(mapped, score.Checkpoint{
			Type:       checkpoint.Type,
			Anchored:   checkpoint.AnchorStatus == domain.Confirmed,
			OccurredAt: checkpoint.OccurredAt,
			ReportedAt: checkpoint.ReportedAt,
			Superseded: checkpoint.SupersededByCheckpointID != nil,
		})
	}
	return mapped
}

func countLiving(checkpoints []domain.Checkpoint) int {
	living := 0
	for _, checkpoint := range checkpoints {
		if checkpoint.SupersededByCheckpointID == nil {
			living++
		}
	}
	return living
}

func optional(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
