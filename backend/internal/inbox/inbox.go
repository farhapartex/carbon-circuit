package inbox

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/domain"
)

var ErrAlreadyProcessed = errors.New("event has already been processed")

type Record struct {
	domain.Base
	EventID       uuid.UUID `gorm:"column:event_id;type:uuid"`
	Topic         string    `gorm:"column:topic"`
	ConsumerGroup string    `gorm:"column:consumer_group"`
	ProcessedAt   time.Time `gorm:"column:processed_at"`
}

func (Record) TableName() string { return "inbox_events" }

type Receipt struct {
	EventID       uuid.UUID
	Topic         string
	ConsumerGroup string
}

func Claim(tx database.Tx, receipt Receipt) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	record := Record{
		EventID:       receipt.EventID,
		Topic:         receipt.Topic,
		ConsumerGroup: receipt.ConsumerGroup,
		ProcessedAt:   time.Now().UTC(),
	}

	result := tx.Session().
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&record)

	if result.Error != nil {
		return fmt.Errorf("claim inbox event: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrAlreadyProcessed
	}

	return nil
}
