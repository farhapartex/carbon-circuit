package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/domain"
)

var ErrClaimAlreadyIssued = errors.New("this claim has already issued credits")

type LedgerStore interface {
	EnsureClass(tx database.Tx, class *domain.CreditClass) error
	RecordIssuance(tx database.Tx, issuance *domain.CreditIssuance) error
	Credit(tx database.Tx, organizationID, classID uuid.UUID, amount string) error
	Holdings(tx database.Tx, organizationID uuid.UUID) ([]domain.Holding, error)
	Issuances(tx database.Tx, organizationID uuid.UUID, limit int) ([]domain.CreditIssuance, error)
	IssuedForClaim(tx database.Tx, claimID uuid.UUID) (domain.CreditIssuance, bool, error)
}

type LedgerRepository struct{}

func NewLedgerRepository() *LedgerRepository { return &LedgerRepository{} }

func (r *LedgerRepository) EnsureClass(tx database.Tx, class *domain.CreditClass) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token_id"}},
			DoNothing: true,
		}).
		Create(class).Error
	if err != nil {
		return fmt.Errorf("ensure credit class: %w", err)
	}

	if class.ID != uuid.Nil {
		return nil
	}

	return tx.Session().
		First(class, "token_id = ?", class.TokenID).Error
}

func (r *LedgerRepository) RecordIssuance(
	tx database.Tx,
	issuance *domain.CreditIssuance,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(issuance).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrClaimAlreadyIssued
	}
	if err != nil {
		return fmt.Errorf("record issuance: %w", err)
	}

	return nil
}

func (r *LedgerRepository) Credit(
	tx database.Tx,
	organizationID, classID uuid.UUID,
	amount string,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	outcome := tx.Session().Exec(`
		INSERT INTO credit_ledger.credit_balances
		  (organization_id, credit_class_id, available)
		VALUES (?, ?, ?)
		ON CONFLICT (organization_id, credit_class_id)
		DO UPDATE SET available = credit_balances.available + EXCLUDED.available,
		              updated_at = now(),
		              version = credit_balances.version + 1`,
		organizationID, classID, amount)
	if outcome.Error != nil {
		return fmt.Errorf("credit holding: %w", outcome.Error)
	}

	if outcome.RowsAffected != 1 {
		return fmt.Errorf("crediting %s to %s changed no holding", amount, organizationID)
	}

	return nil
}

func (r *LedgerRepository) Holdings(
	tx database.Tx,
	organizationID uuid.UUID,
) ([]domain.Holding, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var rows []struct {
		domain.CreditClass
		Available string
		Escrowed  string
		Retired   string
	}

	err := tx.Session().
		Table("credit_ledger.credit_balances AS balance").
		Select(`class.*, balance.available, balance.escrowed, balance.retired`).
		Joins(`JOIN credit_ledger.credit_classes AS class ON class.id = balance.credit_class_id`).
		Where("balance.organization_id = ? AND balance.deleted_at IS NULL", organizationID).
		Order("class.facility_name, class.vintage_year DESC, class.activity_type").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("load holdings: %w", err)
	}

	holdings := make([]domain.Holding, 0, len(rows))
	for _, row := range rows {
		holdings = append(holdings, domain.Holding{
			Class:     row.CreditClass,
			Available: row.Available,
			Escrowed:  row.Escrowed,
			Retired:   row.Retired,
		})
	}

	return holdings, nil
}

func (r *LedgerRepository) Issuances(
	tx database.Tx,
	organizationID uuid.UUID,
	limit int,
) ([]domain.CreditIssuance, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var issuances []domain.CreditIssuance

	err := tx.Session().
		Where("organization_id = ?", organizationID).
		Order("issued_at DESC").
		Limit(limit).
		Find(&issuances).Error
	if err != nil {
		return nil, fmt.Errorf("load issuances: %w", err)
	}

	return issuances, nil
}

func (r *LedgerRepository) IssuedForClaim(
	tx database.Tx,
	claimID uuid.UUID,
) (domain.CreditIssuance, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.CreditIssuance{}, false, err
	}

	var issuance domain.CreditIssuance

	err := tx.Session().First(&issuance, "claim_id = ?", claimID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.CreditIssuance{}, false, nil
	}
	if err != nil {
		return domain.CreditIssuance{}, false, fmt.Errorf("find issuance for claim: %w", err)
	}

	return issuance, true, nil
}
