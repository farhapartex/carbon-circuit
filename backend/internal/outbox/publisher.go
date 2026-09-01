package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/carboncircuit/backend/internal/kafka"
)

type Dispatcher interface {
	Publish(ctx context.Context, messages ...kafka.Message) error
}

type PublisherOptions struct {
	Database  *gorm.DB
	Dispatch  Dispatcher
	Logger    *slog.Logger
	Interval  time.Duration
	BatchSize int
}

type Publisher struct {
	options PublisherOptions
}

func NewPublisher(options PublisherOptions) *Publisher {
	if options.Interval <= 0 {
		options.Interval = time.Second
	}
	if options.BatchSize <= 0 {
		options.BatchSize = 100
	}
	return &Publisher{options: options}
}

func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.options.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if drained, err := p.DrainOnce(ctx); err != nil {
				p.options.Logger.Error("outbox drain failed", slog.Any("error", err))
			} else if drained > 0 {
				p.options.Logger.Info("outbox drained", slog.Int("events", drained))
			}
		}
	}
}

func (p *Publisher) DrainOnce(ctx context.Context) (int, error) {
	published := 0

	err := p.options.Database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var claimed []Event

		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("published_at IS NULL").
			Order("created_at ASC").
			Limit(p.options.BatchSize).
			Find(&claimed).Error
		if err != nil {
			return fmt.Errorf("claim outbox events: %w", err)
		}

		if len(claimed) == 0 {
			return nil
		}

		messages := make([]kafka.Message, 0, len(claimed))
		for _, event := range claimed {
			message, buildErr := messageFrom(event)
			if buildErr != nil {
				return buildErr
			}
			messages = append(messages, message)
		}

		if err := p.options.Dispatch.Publish(ctx, messages...); err != nil {
			return fmt.Errorf("dispatch outbox events: %w", err)
		}

		if err := markPublished(tx, claimed); err != nil {
			return err
		}

		published = len(claimed)
		return nil
	})

	return published, err
}

func messageFrom(event Event) (kafka.Message, error) {
	headers := map[string]string{}
	if len(event.Headers) > 0 {
		if err := json.Unmarshal(event.Headers, &headers); err != nil {
			return kafka.Message{}, fmt.Errorf("decode headers for event %s: %w", event.ID, err)
		}
	}

	headers["event_id"] = event.ID.String()
	headers["aggregate_type"] = event.AggregateType

	return kafka.Message{
		Topic:   event.EventType,
		Key:     []byte(event.AggregateID.String()),
		Value:   event.Payload,
		Headers: headers,
	}, nil
}

func markPublished(tx *gorm.DB, events []Event) error {
	ids := make([]uuid.UUID, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}

	publishedAt := time.Now()

	result := tx.Model(&Event{}).
		Where("id IN ? AND published_at IS NULL", ids).
		Updates(map[string]any{"published_at": publishedAt, "updated_at": publishedAt})

	if result.Error != nil {
		return fmt.Errorf("mark outbox events published: %w", result.Error)
	}

	if result.RowsAffected != int64(len(ids)) {
		return errors.New("outbox events changed while being published")
	}

	return nil
}
