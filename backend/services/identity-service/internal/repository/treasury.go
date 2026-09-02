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

var (
	ErrNonceUnknown      = errors.New("nonce was not issued or has already been used")
	ErrTreasuryExists    = errors.New("organization already has an active treasury address")
	ErrAddressDesignated = errors.New("address is already the treasury of another organization")
)

type TreasuryStore interface {
	IssueNonce(tx database.Tx, nonce *domain.SiweNonce) error
	ConsumeNonce(tx database.Tx, nonce string, at time.Time) (domain.SiweNonce, error)
	InsertTreasury(tx database.Tx, treasury *domain.TreasuryAddress) error
}

type TreasuryRepository struct{}

func NewTreasuryRepository() *TreasuryRepository { return &TreasuryRepository{} }

func (r *TreasuryRepository) IssueNonce(tx database.Tx, nonce *domain.SiweNonce) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	if err := tx.Session().Create(nonce).Error; err != nil {
		return fmt.Errorf("issue siwe nonce: %w", err)
	}

	return nil
}

func (r *TreasuryRepository) ConsumeNonce(
	tx database.Tx,
	nonce string,
	at time.Time,
) (domain.SiweNonce, error) {
	if err := tx.Bound(); err != nil {
		return domain.SiweNonce{}, err
	}

	result := tx.Session().Model(&domain.SiweNonce{}).
		Where("nonce = ? AND consumed_at IS NULL AND expires_at > ?", nonce, at).
		Updates(map[string]any{"consumed_at": at, "updated_at": at})

	if result.Error != nil {
		return domain.SiweNonce{}, fmt.Errorf("consume siwe nonce: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.SiweNonce{}, ErrNonceUnknown
	}

	var consumed domain.SiweNonce
	if err := tx.Session().First(&consumed, "nonce = ?", nonce).Error; err != nil {
		return domain.SiweNonce{}, fmt.Errorf("load consumed nonce: %w", err)
	}

	return consumed, nil
}

func (r *TreasuryRepository) InsertTreasury(
	tx database.Tx,
	treasury *domain.TreasuryAddress,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(treasury).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return classifyConflict(tx, treasury.OrganizationID, treasury.Address)
	}
	if err != nil {
		return fmt.Errorf("insert treasury address: %w", err)
	}

	return nil
}

func classifyConflict(tx database.Tx, organizationID uuid.UUID, address string) error {
	var mine int64
	tx.Session().Model(&domain.TreasuryAddress{}).
		Where("organization_id = ? AND state = ?", organizationID, domain.TreasuryActive).
		Count(&mine)

	if mine > 0 {
		return ErrTreasuryExists
	}

	_ = address
	return ErrAddressDesignated
}
