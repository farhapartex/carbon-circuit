package idempotency

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"github.com/carboncircuit/backend/internal/database"
)

var (
	ErrInProgress = errors.New("an identical request is already being processed")
	ErrKeyReused  = errors.New("idempotency key was used with a different request")
)

type Reservation struct {
	RecordID uuid.UUID
	Replay   *Response
}

func (r Reservation) IsReplay() bool { return r.Replay != nil }

func Reserve(tx database.Tx, request Request) (Reservation, error) {
	if err := tx.Bound(); err != nil {
		return Reservation{}, err
	}

	record := Record{
		OrganizationID: request.Scope.OrganizationID,
		UserID:         request.Scope.UserID,
		Endpoint:       request.Endpoint,
		IdempotencyKey: request.Key,
		RequestHash:    request.hash(),
		State:          StateProcessing,
	}

	result := tx.Session().Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
	if result.Error != nil {
		return Reservation{}, fmt.Errorf("reserve idempotency key: %w", result.Error)
	}

	if result.RowsAffected == 1 {
		return Reservation{RecordID: record.ID}, nil
	}

	return adopt(tx, request)
}

func adopt(tx database.Tx, request Request) (Reservation, error) {
	var existing Record

	query := tx.Session().
		Where("endpoint = ? AND idempotency_key = ?", request.Endpoint, request.Key)

	if request.Scope.OrganizationID != nil {
		query = query.Where("organization_id = ?", request.Scope.OrganizationID)
	} else {
		query = query.Where("organization_id IS NULL AND user_id = ?", request.Scope.UserID)
	}

	if err := query.First(&existing).Error; err != nil {
		return Reservation{}, fmt.Errorf("load reserved idempotency key: %w", err)
	}

	if !bytes.Equal(existing.RequestHash, request.hash()) {
		return Reservation{}, ErrKeyReused
	}

	switch existing.State {
	case StateCompleted:
		return Reservation{RecordID: existing.ID, Replay: replayOf(existing)}, nil
	case StateFailed:
		return retake(tx, existing)
	default:
		return Reservation{}, ErrInProgress
	}
}

func replayOf(record Record) *Response {
	response := Response{Body: record.ResponseBody.Bytes(), ResourceID: record.ResourceID}
	if record.ResponseStatus != nil {
		response.Status = *record.ResponseStatus
	}
	return &response
}

func retake(tx database.Tx, existing Record) (Reservation, error) {
	result := tx.Session().Model(&Record{}).
		Where("id = ? AND state = ?", existing.ID, StateFailed).
		Updates(map[string]any{"state": StateProcessing, "updated_at": time.Now()})

	if result.Error != nil {
		return Reservation{}, fmt.Errorf("retake failed idempotency key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return Reservation{}, ErrInProgress
	}

	return Reservation{RecordID: existing.ID}, nil
}

func Complete(tx database.Tx, recordID uuid.UUID, response Response) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	completedAt := time.Now()

	result := tx.Session().Model(&Record{}).
		Where("id = ? AND state = ?", recordID, StateProcessing).
		Updates(map[string]any{
			"state":           StateCompleted,
			"response_status": response.Status,
			"response_body":   database.JSONDocument(response.Body),
			"resource_id":     response.ResourceID,
			"completed_at":    completedAt,
			"updated_at":      completedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("complete idempotency record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("complete idempotency record %s: no longer processing", recordID)
	}

	return nil
}

func Fail(tx database.Tx, recordID uuid.UUID) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Model(&Record{}).
		Where("id = ? AND state = ?", recordID, StateProcessing).
		Updates(map[string]any{"state": StateFailed, "updated_at": time.Now()}).Error
	if err != nil {
		return fmt.Errorf("fail idempotency record: %w", err)
	}

	return nil
}
