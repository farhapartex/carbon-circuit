package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-service/internal/repository"
)

const (
	logCheckpointEndpoint = "POST /v1/batches/:batchId/checkpoints"
	checkpointLoggedEvent = "checkpoint.logged"
)

var (
	ErrCheckpointBeforeProduction = errors.New("a checkpoint cannot precede the batch's production date")
	ErrCheckpointInFuture         = errors.New("a checkpoint cannot be logged for a future time")
	ErrShippingMethodRequired     = errors.New("a movement needs a shipping method")
	ErrNotBatchOwner              = errors.New("only the organization that owns this batch may log checkpoints on it")
	ErrCorrectionWindowClosed     = errors.New("the seven day correction window has closed")
	ErrCheckpointNotFound         = errors.New("checkpoint not found")
	ErrAlreadyCorrected           = errors.New("this checkpoint has already been corrected")
	ErrCorrectionReasonRequired   = errors.New("a correction needs a reason")
)

var checkpointTypes = map[domain.CheckpointType]struct{}{
	domain.ProductionComplete: {},
	domain.DepartedOrigin:     {},
	domain.CustomsExport:      {},
	domain.CustomsImport:      {},
	domain.ArrivedDestination: {},
}

var shippingMethods = map[string]struct{}{
	"air_freight_short_haul": {}, "air_freight_long_haul": {},
	"sea_freight_container": {}, "sea_freight_bulk": {},
	"rail_electric": {}, "rail_diesel": {},
	"road_hgv": {}, "road_lgv": {}, "inland_waterway": {},
}

type CheckpointEntry struct {
	Type             domain.CheckpointType
	LocationLabel    string
	CountryCode      string
	Latitude         string
	Longitude        string
	ShippingMethod   string
	OccurredAt       time.Time
	CorrectsID       *uuid.UUID
	CorrectionReason string
	ExternalID       string
	IdempotencyKey   string
	RequestBody      []byte
}

func (e CheckpointEntry) Validate(now time.Time) error {
	if _, known := checkpointTypes[e.Type]; !known {
		return fmt.Errorf("checkpoint type %q is not a known type", e.Type)
	}
	if strings.TrimSpace(e.LocationLabel) == "" {
		return fmt.Errorf("location is required")
	}
	if len(e.CountryCode) != 2 {
		return fmt.Errorf("country must be a two letter code")
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("event time is required")
	}
	if e.OccurredAt.After(now) {
		return ErrCheckpointInFuture
	}
	if e.Type != domain.ProductionComplete {
		if _, known := shippingMethods[e.ShippingMethod]; !known {
			return ErrShippingMethodRequired
		}
	}
	if (e.Latitude == "") != (e.Longitude == "") {
		return fmt.Errorf("coordinates must be given as a pair or omitted")
	}
	if e.CorrectsID != nil && strings.TrimSpace(e.CorrectionReason) == "" {
		return ErrCorrectionReasonRequired
	}
	return nil
}

type LoggedCheckpoint struct {
	Checkpoint domain.Checkpoint
	Batch      domain.Batch
	Replayed   bool
}

func (s *BatchService) LogCheckpoint(
	ctx context.Context,
	actor Actor,
	batchID uuid.UUID,
	entry CheckpointEntry,
) (LoggedCheckpoint, error) {
	if err := actor.mayProduce(); err != nil {
		return LoggedCheckpoint{}, err
	}

	now := time.Now().UTC()
	if err := entry.Validate(now); err != nil {
		return LoggedCheckpoint{}, err
	}

	checkpointID, err := uuid.NewV7()
	if err != nil {
		return LoggedCheckpoint{}, fmt.Errorf("generate checkpoint id: %w", err)
	}

	var logged LoggedCheckpoint

	err = s.tenant(ctx, actor, func(tx database.Tx) error {
		recorded, replay, workErr := s.appendCheckpoint(
			tx, actor, batchID, checkpointID, entry, now,
		)
		if workErr != nil {
			return workErr
		}
		logged = recorded
		logged.Replayed = replay
		return nil
	})
	if err != nil {
		return LoggedCheckpoint{}, err
	}

	return logged, nil
}

func (s *BatchService) appendCheckpoint(
	tx database.Tx,
	actor Actor,
	batchID, checkpointID uuid.UUID,
	entry CheckpointEntry,
	now time.Time,
) (LoggedCheckpoint, bool, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForOrganization(actor.OrganizationID),
		Endpoint: logCheckpointEndpoint,
		Key:      entry.IdempotencyKey,
		Body:     entry.RequestBody,
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return LoggedCheckpoint{}, false, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return LoggedCheckpoint{}, false, ErrIdempotencyConflict
	case err != nil:
		return LoggedCheckpoint{}, false, err
	}

	if reservation.IsReplay() {
		var replayed LoggedCheckpoint
		if err := json.Unmarshal(reservation.Replay.Body, &replayed); err != nil {
			return LoggedCheckpoint{}, false, fmt.Errorf("decode replayed checkpoint: %w", err)
		}
		return replayed, true, nil
	}

	batch, found, err := s.batches.FindOwned(tx, actor.OrganizationID, batchID)
	if err != nil {
		return LoggedCheckpoint{}, false, err
	}
	if !found {
		if _, visible, findErr := s.batches.Find(tx, batchID); findErr != nil {
			return LoggedCheckpoint{}, false, findErr
		} else if visible {
			return LoggedCheckpoint{}, false, ErrNotBatchOwner
		}
		return LoggedCheckpoint{}, false, ErrBatchNotFound
	}

	if entry.OccurredAt.Before(batch.ProducedAt) {
		return LoggedCheckpoint{}, false, ErrCheckpointBeforeProduction
	}

	if entry.CorrectsID != nil {
		if err := s.verifyCorrectable(tx, batchID, *entry.CorrectsID, now); err != nil {
			return LoggedCheckpoint{}, false, err
		}
	}

	checkpoint := domain.Checkpoint{
		OrganizationID:             batch.OrganizationID,
		BatchID:                    batchID,
		Type:                       entry.Type,
		LocationLabel:              strings.TrimSpace(entry.LocationLabel),
		CountryCode:                strings.ToUpper(entry.CountryCode),
		Latitude:                   optional(entry.Latitude),
		Longitude:                  optional(entry.Longitude),
		ShippingMethod:             optional(entry.ShippingMethod),
		OccurredAt:                 entry.OccurredAt,
		ReportedAt:                 now,
		ReportedByOrganizationID:   actor.OrganizationID,
		ReportedByOrganizationName: actor.OrganizationName,
		AnchorStatus:               domain.Unanchored,
		SupersedesCheckpointID:     entry.CorrectsID,
		CorrectionReason:           optional(entry.CorrectionReason),
		ExternalID:                 optional(entry.ExternalID),
	}
	checkpoint.ID = checkpointID

	if err := s.checkpoints.Insert(tx, &checkpoint); err != nil {
		return LoggedCheckpoint{}, false, err
	}

	if entry.CorrectsID != nil {
		if err := s.checkpoints.MarkSuperseded(tx, *entry.CorrectsID, checkpointID); err != nil {
			if errors.Is(err, repository.ErrAlreadyCorrected) {
				return LoggedCheckpoint{}, false, ErrAlreadyCorrected
			}
			return LoggedCheckpoint{}, false, err
		}
	}

	declarations, err := s.batches.ListParents(tx, batchID)
	if err != nil {
		return LoggedCheckpoint{}, false, err
	}

	parents := make([]ParentView, 0, len(declarations))
	for _, declaration := range declarations {
		parents = append(parents, ParentView{
			DeclaredReference: declaration.DeclaredReference,
			Resolved:          declaration.ParentBatchID != nil,
		})
	}

	if err := s.rescore(tx, &batch, parents); err != nil {
		return LoggedCheckpoint{}, false, err
	}

	if _, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: batchAggregate,
		AggregateID:   batchID,
		EventType:     checkpointLoggedEvent,
		Payload: map[string]string{
			"batch_id":        batchID.String(),
			"checkpoint_id":   checkpoint.ID.String(),
			"organization_id": batch.OrganizationID.String(),
			"type":            string(checkpoint.Type),
		},
	}); err != nil {
		return LoggedCheckpoint{}, false, err
	}

	logged := LoggedCheckpoint{Checkpoint: checkpoint, Batch: batch}

	body, err := json.Marshal(logged)
	if err != nil {
		return LoggedCheckpoint{}, false, fmt.Errorf("encode idempotent response: %w", err)
	}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     201,
		Body:       body,
		ResourceID: &checkpoint.ID,
	}); err != nil {
		return LoggedCheckpoint{}, false, err
	}

	return logged, false, nil
}

func (s *BatchService) verifyCorrectable(
	tx database.Tx,
	batchID, originalID uuid.UUID,
	now time.Time,
) error {
	original, found, err := s.checkpoints.Find(tx, originalID)
	if err != nil {
		return err
	}
	if !found || original.BatchID != batchID {
		return ErrCheckpointNotFound
	}
	if original.SupersededByCheckpointID != nil {
		return ErrAlreadyCorrected
	}
	if now.Sub(original.ReportedAt) > domain.CorrectionWindowDays*24*time.Hour {
		return ErrCorrectionWindowClosed
	}
	return nil
}
