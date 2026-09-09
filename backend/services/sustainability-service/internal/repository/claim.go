package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
)

type ClaimStore interface {
	Insert(tx database.Tx, claim *domain.Claim) error
	AttachEvidence(tx database.Tx, attachments []domain.ClaimEvidence) error
	Find(tx database.Tx, organizationID, claimID uuid.UUID) (domain.Claim, bool, error)
	List(tx database.Tx, organizationID uuid.UUID, status domain.ClaimStatus, after string, limit int) ([]domain.Claim, error)
	Evidence(tx database.Tx, organizationID, claimID uuid.UUID) ([]domain.ClaimEvidence, error)
	ConsumedForVintage(tx database.Tx, organizationID, facilityID uuid.UUID, vintageYear int, activity domain.ActivityType) (string, error)
}

type ReferenceStore interface {
	Effective(tx database.Tx, kind domain.FactorKind, lookupKey string, on time.Time) (domain.ReferenceFactor, bool, error)
}

type ClaimRepository struct{}

func NewClaimRepository() *ClaimRepository { return &ClaimRepository{} }

func (r *ClaimRepository) Insert(tx database.Tx, claim *domain.Claim) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	if err := tx.Session().Create(claim).Error; err != nil {
		return fmt.Errorf("insert claim: %w", err)
	}

	return nil
}

func (r *ClaimRepository) AttachEvidence(
	tx database.Tx,
	attachments []domain.ClaimEvidence,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}
	if len(attachments) == 0 {
		return nil
	}

	if err := tx.Session().Create(&attachments).Error; err != nil {
		return fmt.Errorf("attach claim evidence: %w", err)
	}

	return nil
}

func (r *ClaimRepository) Find(
	tx database.Tx,
	organizationID, claimID uuid.UUID,
) (domain.Claim, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Claim{}, false, err
	}

	var claim domain.Claim

	err := tx.Session().
		First(&claim, "id = ? AND organization_id = ?", claimID, organizationID).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Claim{}, false, nil
	}
	if err != nil {
		return domain.Claim{}, false, fmt.Errorf("find claim: %w", err)
	}

	return claim, true, nil
}

func (r *ClaimRepository) List(
	tx database.Tx,
	organizationID uuid.UUID,
	status domain.ClaimStatus,
	after string,
	limit int,
) ([]domain.Claim, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	query := tx.Session().
		Where("organization_id = ?", organizationID).
		Order("created_at DESC, id DESC").
		Limit(limit)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if after != "" {
		query = query.Where("id < ?", after)
	}

	var claims []domain.Claim
	if err := query.Find(&claims).Error; err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}

	return claims, nil
}

func (r *ClaimRepository) Evidence(
	tx database.Tx,
	organizationID, claimID uuid.UUID,
) ([]domain.ClaimEvidence, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var attachments []domain.ClaimEvidence

	err := tx.Session().
		Where("organization_id = ? AND claim_id = ?", organizationID, claimID).
		Order("created_at").
		Find(&attachments).Error
	if err != nil {
		return nil, fmt.Errorf("load claim evidence: %w", err)
	}

	return attachments, nil
}

func (r *ClaimRepository) ConsumedForVintage(
	tx database.Tx,
	organizationID, facilityID uuid.UUID,
	vintageYear int,
	activity domain.ActivityType,
) (string, error) {
	if err := tx.Bound(); err != nil {
		return "0", err
	}

	var total *string

	err := tx.Session().
		Model(&domain.Claim{}).
		Select(`coalesce(sum(
			CASE
				WHEN status = 'approved' THEN coalesce(issued_amount, 0)
				ELSE least(requested_amount, computed_ceiling)
			END
		), 0)::text`).
		Where("organization_id = ? AND facility_id = ? AND vintage_year = ? AND activity_type = ?",
			organizationID, facilityID, vintageYear, activity).
		Where("status <> ?", domain.Rejected).
		Where("deleted_at IS NULL").
		Scan(&total).Error
	if err != nil {
		return "0", fmt.Errorf("sum consumed ceiling: %w", err)
	}

	if total == nil {
		return "0", nil
	}

	return *total, nil
}

type ReferenceRepository struct{}

func NewReferenceRepository() *ReferenceRepository { return &ReferenceRepository{} }

func (r *ReferenceRepository) Effective(
	tx database.Tx,
	kind domain.FactorKind,
	lookupKey string,
	on time.Time,
) (domain.ReferenceFactor, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.ReferenceFactor{}, false, err
	}

	var factor domain.ReferenceFactor

	err := tx.Session().
		Where("kind = ? AND lookup_key = ?", kind, lookupKey).
		Where("effective_from <= ?", on).
		Where("effective_to IS NULL OR effective_to > ?", on).
		Order("effective_from DESC").
		First(&factor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ReferenceFactor{}, false, nil
	}
	if err != nil {
		return domain.ReferenceFactor{}, false, fmt.Errorf("find reference factor: %w", err)
	}

	return factor, true, nil
}
