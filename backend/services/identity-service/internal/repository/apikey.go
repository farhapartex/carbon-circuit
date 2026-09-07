package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

type APIKeyStore interface {
	Insert(tx database.Tx, key *domain.APIKey) error
	List(tx database.Tx, organizationID uuid.UUID) ([]domain.APIKey, error)
	CountActive(tx database.Tx, organizationID uuid.UUID) (int, error)
	Revoke(
		tx database.Tx,
		organizationID, keyID, revokedBy uuid.UUID,
	) (bool, error)
}

type APIKeyRepository struct{}

func NewAPIKeyRepository() *APIKeyRepository { return &APIKeyRepository{} }

func (r *APIKeyRepository) Insert(tx database.Tx, key *domain.APIKey) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(key).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrPrefixTaken
	}
	if err != nil {
		return fmt.Errorf("insert api key: %w", err)
	}

	return nil
}

func (r *APIKeyRepository) List(
	tx database.Tx,
	organizationID uuid.UUID,
) ([]domain.APIKey, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var keys []domain.APIKey
	err := tx.Session().
		Where("organization_id = ?", organizationID).
		Order("revoked_at IS NOT NULL, created_at DESC").
		Find(&keys).Error
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}

	return keys, nil
}

func (r *APIKeyRepository) CountActive(
	tx database.Tx,
	organizationID uuid.UUID,
) (int, error) {
	if err := tx.Bound(); err != nil {
		return 0, err
	}

	var total int64
	err := tx.Session().Model(&domain.APIKey{}).
		Where("organization_id = ? AND revoked_at IS NULL", organizationID).
		Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("count active api keys: %w", err)
	}

	return int(total), nil
}

func (r *APIKeyRepository) Revoke(
	tx database.Tx,
	organizationID, keyID, revokedBy uuid.UUID,
) (bool, error) {
	if err := tx.Bound(); err != nil {
		return false, err
	}

	result := tx.Session().Model(&domain.APIKey{}).
		Where(
			"id = ? AND organization_id = ? AND revoked_at IS NULL",
			keyID, organizationID,
		).
		Updates(map[string]any{
			"revoked_at":         time.Now().UTC(),
			"revoked_by_user_id": revokedBy,
			"updated_at":         gorm.Expr("now()"),
			"version":            gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return false, fmt.Errorf("revoke api key: %w", result.Error)
	}

	return result.RowsAffected > 0, nil
}

var ErrPrefixTaken = errors.New("api key prefix already exists")
