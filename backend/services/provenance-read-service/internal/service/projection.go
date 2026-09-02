package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/inbox"
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/repository"
)

var ErrUnknownTopic = errors.New("no projection for this topic")

type Projection struct {
	database      *gorm.DB
	store         repository.PublicStore
	consumerGroup string
	logger        *slog.Logger
}

func NewProjection(
	handle *gorm.DB,
	store repository.PublicStore,
	consumerGroup string,
	logger *slog.Logger,
) *Projection {
	return &Projection{
		database:      handle,
		store:         store,
		consumerGroup: consumerGroup,
		logger:        logger,
	}
}

func (p *Projection) Apply(ctx context.Context, delivery kafka.Delivery) error {
	eventID, err := uuid.Parse(delivery.EventID)
	if err != nil {
		return fmt.Errorf("event %q carries no usable event id", delivery.Topic)
	}

	return database.Within(ctx, p.database, func(tx database.Tx) error {
		claimErr := inbox.Claim(tx, inbox.Receipt{
			EventID:       eventID,
			Topic:         delivery.Topic,
			ConsumerGroup: p.consumerGroup,
		})
		if errors.Is(claimErr, inbox.ErrAlreadyProcessed) {
			return nil
		}
		if claimErr != nil {
			return claimErr
		}

		switch delivery.Topic {
		case events.TopicBatchCreated:
			return p.applyBatchCreated(tx, delivery.Value)
		case events.TopicCheckpointLogged:
			return p.applyCheckpointLogged(tx, delivery.Value)
		default:
			return fmt.Errorf("%w: %s", ErrUnknownTopic, delivery.Topic)
		}
	})
}

func (p *Projection) applyBatchCreated(tx database.Tx, payload []byte) error {
	var created events.BatchCreated
	if err := json.Unmarshal(payload, &created); err != nil {
		return fmt.Errorf("decode batch.created: %w", err)
	}

	batchID, err := uuid.Parse(created.BatchID)
	if err != nil {
		return fmt.Errorf("batch.created carries an unusable batch id: %w", err)
	}

	producedAt, err := time.Parse(time.RFC3339, created.ProducedAt)
	if err != nil {
		return fmt.Errorf("batch.created carries an unusable production date: %w", err)
	}

	components, err := json.Marshal(created.ScoreComponents)
	if err != nil {
		return fmt.Errorf("encode score components: %w", err)
	}

	batch := domain.PublicBatch{
		PublicReference:            created.PublicReference,
		ProductCategory:            created.ProductCategory,
		ComponentType:              created.ComponentType,
		OriginatingFacilityName:    created.OriginatingFacilityName,
		OriginatingFacilityCountry: created.OriginatingFacilityCountry,
		ProducedAt:                 producedAt,
		ProvenanceScore:            created.ProvenanceScore,
		ScoreComponents:            components,
		LastUpdatedAt:              time.Now().UTC(),
		Projected:                  true,
	}
	batch.ID = batchID

	return p.store.UpsertBatch(tx, &batch)
}

func (p *Projection) applyCheckpointLogged(tx database.Tx, payload []byte) error {
	var logged events.CheckpointLogged
	if err := json.Unmarshal(payload, &logged); err != nil {
		return fmt.Errorf("decode checkpoint.logged: %w", err)
	}

	batchID, err := uuid.Parse(logged.BatchID)
	if err != nil {
		return fmt.Errorf("checkpoint.logged carries an unusable batch id: %w", err)
	}

	checkpointID, err := uuid.Parse(logged.CheckpointID)
	if err != nil {
		return fmt.Errorf("checkpoint.logged carries an unusable checkpoint id: %w", err)
	}

	occurredAt, err := time.Parse(time.RFC3339, logged.OccurredAt)
	if err != nil {
		return fmt.Errorf("checkpoint.logged carries an unusable event time: %w", err)
	}

	if err := p.store.EnsureBatch(tx, batchID); err != nil {
		return err
	}

	checkpoint := domain.PublicCheckpoint{
		BatchID:       batchID,
		Type:          logged.Type,
		LocationLabel: logged.LocationLabel,
		CountryCode:   logged.CountryCode,
		OccurredAt:    occurredAt,
		AnchorStatus:  logged.AnchorStatus,
	}
	checkpoint.ID = checkpointID

	if logged.ShippingMethod != "" {
		method := logged.ShippingMethod
		checkpoint.ShippingMethod = &method
	}

	if err := p.store.UpsertCheckpoint(tx, &checkpoint); err != nil {
		return err
	}

	if logged.SupersedesCheckpointID != "" {
		superseded, parseErr := uuid.Parse(logged.SupersedesCheckpointID)
		if parseErr != nil {
			return fmt.Errorf("checkpoint.logged names an unusable superseded id: %w", parseErr)
		}
		if err := p.store.MarkSuperseded(tx, superseded); err != nil {
			return err
		}
	}

	components, err := json.Marshal(logged.ScoreComponents)
	if err != nil {
		return fmt.Errorf("encode score components: %w", err)
	}

	return p.store.ApplyScore(tx, batchID, logged.ProvenanceScore, components)
}
