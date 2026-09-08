package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/apikey"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

var (
	ErrAPIKeyNotFound  = errors.New("api key not found")
	ErrAPIKeyNameEmpty = errors.New("name the key after the system that will use it")
)

const maximumAPIKeyNameLength = 80

type IssuedAPIKey struct {
	Key       domain.APIKey
	Presented string
}

type APIKeyService struct {
	database      *gorm.DB
	keys          repository.APIKeyStore
	organizations repository.OrganizationReader
	hasher        *apikey.Hasher
	logger        *slog.Logger
}

func NewAPIKeyService(
	handle *gorm.DB,
	keys repository.APIKeyStore,
	organizations repository.OrganizationReader,
	hasher *apikey.Hasher,
	logger *slog.Logger,
) *APIKeyService {
	return &APIKeyService{
		database:      handle,
		keys:          keys,
		organizations: organizations,
		hasher:        hasher,
		logger:        logger,
	}
}

func (s *APIKeyService) Create(
	ctx context.Context,
	actor Actor,
	name string,
) (IssuedAPIKey, error) {
	if !actor.manages() {
		return IssuedAPIKey{}, ErrNotPermitted
	}

	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > maximumAPIKeyNameLength {
		return IssuedAPIKey{}, ErrAPIKeyNameEmpty
	}

	issued, err := s.hasher.Issue()
	if err != nil {
		return IssuedAPIKey{}, err
	}

	keyID, err := uuid.NewV7()
	if err != nil {
		return IssuedAPIKey{}, fmt.Errorf("generate api key id: %w", err)
	}

	key := domain.APIKey{
		OrganizationID:  actor.OrganizationID,
		Name:            trimmed,
		Prefix:          issued.Prefix,
		SecretHMAC:      issued.HMAC,
		CreatedByUserID: actor.UserID,
	}
	key.ID = keyID

	err = s.tenant(ctx, actor, func(tx database.Tx) error {
		return s.keys.Insert(tx, &key)
	})
	if err != nil {
		return IssuedAPIKey{}, err
	}

	return IssuedAPIKey{Key: key, Presented: issued.Presented()}, nil
}

func (s *APIKeyService) List(
	ctx context.Context,
	actor Actor,
) ([]domain.APIKey, error) {
	var keys []domain.APIKey

	err := s.tenant(ctx, actor, func(tx database.Tx) error {
		listed, err := s.keys.List(tx, actor.OrganizationID)
		if err != nil {
			return err
		}
		keys = listed
		return nil
	})

	return keys, err
}

func (s *APIKeyService) Revoke(
	ctx context.Context,
	actor Actor,
	keyID uuid.UUID,
) error {
	if !actor.manages() {
		return ErrNotPermitted
	}

	return s.tenant(ctx, actor, func(tx database.Tx) error {
		revoked, err := s.keys.Revoke(tx, actor.OrganizationID, keyID, actor.UserID)
		if err != nil {
			return err
		}
		if !revoked {
			return ErrAPIKeyNotFound
		}
		return nil
	})
}

func (s *APIKeyService) tenant(
	ctx context.Context,
	actor Actor,
	work func(tx database.Tx) error,
) error {
	return database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         actor.UserID.String(),
			OrganizationID: actor.OrganizationID.String(),
		},
		work,
	)
}
