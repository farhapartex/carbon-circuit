package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionSighting struct {
	UserID         uuid.UUID
	Auth0SessionID string
	UserAgent      string
	IPAddress      string
}

type SessionRegistry struct {
	database *gorm.DB
	sessions repository.SessionStore
	logger   *slog.Logger
}

func NewSessionRegistry(
	handle *gorm.DB,
	sessions repository.SessionStore,
	logger *slog.Logger,
) *SessionRegistry {
	return &SessionRegistry{database: handle, sessions: sessions, logger: logger}
}

func (r *SessionRegistry) Record(
	ctx context.Context,
	sighting SessionSighting,
) error {
	if sighting.Auth0SessionID == "" {
		return nil
	}

	now := time.Now().UTC()

	session := domain.Session{
		UserID:         sighting.UserID,
		Auth0SessionID: sighting.Auth0SessionID,
		UserAgent:      sighting.UserAgent,
		IPAddress:      sighting.IPAddress,
		StartedAt:      now,
		LastSeenAt:     now,
	}

	return database.WithinTenant(ctx, r.database,
		database.TenantContext{UserID: sighting.UserID.String()},
		func(tx database.Tx) error {
			return r.sessions.Record(tx, &session)
		},
	)
}

func (r *SessionRegistry) List(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.Session, error) {
	var sessions []domain.Session

	err := database.WithinTenant(ctx, r.database,
		database.TenantContext{UserID: userID.String()},
		func(tx database.Tx) error {
			listed, err := r.sessions.ListActive(tx, userID)
			if err != nil {
				return err
			}
			sessions = listed
			return nil
		},
	)

	return sessions, err
}

func (r *SessionRegistry) Revoke(
	ctx context.Context,
	userID uuid.UUID,
	auth0SessionID string,
) error {
	return database.WithinTenant(ctx, r.database,
		database.TenantContext{UserID: userID.String()},
		func(tx database.Tx) error {
			revoked, err := r.sessions.Revoke(tx, userID, auth0SessionID)
			if err != nil {
				return err
			}
			if !revoked {
				return ErrSessionNotFound
			}
			return nil
		},
	)
}
