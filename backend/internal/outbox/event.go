package outbox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
)

type Event struct {
	ID            uuid.UUID             `gorm:"column:id;type:uuid;primaryKey;default:uuidv7()"`
	CreatedAt     time.Time             `gorm:"column:created_at"`
	UpdatedAt     time.Time             `gorm:"column:updated_at"`
	Version       int                   `gorm:"column:version;default:1"`
	AggregateType string                `gorm:"column:aggregate_type"`
	AggregateID   uuid.UUID             `gorm:"column:aggregate_id;type:uuid"`
	EventType     string                `gorm:"column:event_type"`
	Payload       database.JSONDocument `gorm:"column:payload;type:jsonb"`
	Headers       database.JSONDocument `gorm:"column:headers;type:jsonb"`
	PublishedAt   *time.Time            `gorm:"column:published_at"`
}

func (Event) TableName() string { return "outbox_events" }

type Envelope struct {
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       any
	Headers       map[string]string
}

func Append(tx database.Tx, envelope Envelope) (uuid.UUID, error) {
	if err := tx.Bound(); err != nil {
		return uuid.Nil, err
	}

	if envelope.EventType == "" || envelope.AggregateType == "" {
		return uuid.Nil, fmt.Errorf("outbox event requires an aggregate type and event type")
	}

	if envelope.AggregateID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("outbox event requires an aggregate id to key its partition")
	}

	payload, err := json.Marshal(envelope.Payload)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encode outbox payload: %w", err)
	}

	headers, err := json.Marshal(orEmpty(envelope.Headers))
	if err != nil {
		return uuid.Nil, fmt.Errorf("encode outbox headers: %w", err)
	}

	event := Event{
		AggregateType: envelope.AggregateType,
		AggregateID:   envelope.AggregateID,
		EventType:     envelope.EventType,
		Payload:       payload,
		Headers:       headers,
	}

	if err := tx.Session().Create(&event).Error; err != nil {
		return uuid.Nil, fmt.Errorf("append outbox event: %w", err)
	}

	return event.ID, nil
}

func orEmpty(headers map[string]string) map[string]string {
	if headers == nil {
		return map[string]string{}
	}
	return headers
}
