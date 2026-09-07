package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

type SessionStore interface {
	Record(tx database.Tx, session *domain.Session) error
	ListActive(tx database.Tx, userID uuid.UUID) ([]domain.Session, error)
	Revoke(tx database.Tx, userID uuid.UUID, auth0SessionID string) (bool, error)
}

type SessionRepository struct{}

func NewSessionRepository() *SessionRepository { return &SessionRepository{} }

func (r *SessionRepository) Record(
	tx database.Tx,
	session *domain.Session,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"}, {Name: "auth0_session_id"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"last_seen_at": session.LastSeenAt,
				"user_agent":   session.UserAgent,
				"ip_address":   session.IPAddress,
				"updated_at":   gorm.Expr("now()"),
				"version":      gorm.Expr("sessions.version + 1"),
			}),
		}).
		Create(session).Error
	if err != nil {
		return fmt.Errorf("record session: %w", err)
	}

	return nil
}

func (r *SessionRepository) ListActive(
	tx database.Tx,
	userID uuid.UUID,
) ([]domain.Session, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var sessions []domain.Session
	err := tx.Session().
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	return sessions, nil
}

func (r *SessionRepository) Revoke(
	tx database.Tx,
	userID uuid.UUID,
	auth0SessionID string,
) (bool, error) {
	if err := tx.Bound(); err != nil {
		return false, err
	}

	result := tx.Session().Model(&domain.Session{}).
		Where(
			"user_id = ? AND auth0_session_id = ? AND revoked_at IS NULL",
			userID, auth0SessionID,
		).
		Updates(map[string]any{
			"revoked_at": time.Now().UTC(),
			"updated_at": gorm.Expr("now()"),
			"version":    gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return false, fmt.Errorf("revoke session: %w", result.Error)
	}

	return result.RowsAffected > 0, nil
}
