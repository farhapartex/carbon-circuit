package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

const (
	resolvingOrganization = "app.resolving_organization_id"
	resolvingFacility     = "app.resolving_facility_id"
)

type ResolutionStore interface {
	Organization(tx database.Tx, organizationID uuid.UUID) (domain.Organization, bool, error)
	TreasuryAddress(tx database.Tx, organizationID uuid.UUID) (string, error)
	Facility(tx database.Tx, facilityID uuid.UUID) (domain.Facility, bool, error)
}

type ResolutionRepository struct{}

func NewResolutionRepository() *ResolutionRepository { return &ResolutionRepository{} }

func grant(tx database.Tx, setting string, value uuid.UUID) error {
	return tx.Session().
		Exec("SELECT set_config(?, ?, true)", setting, value.String()).Error
}

func (r *ResolutionRepository) Organization(
	tx database.Tx,
	organizationID uuid.UUID,
) (domain.Organization, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Organization{}, false, err
	}
	if err := grant(tx, resolvingOrganization, organizationID); err != nil {
		return domain.Organization{}, false, fmt.Errorf("grant organization resolution: %w", err)
	}

	var organization domain.Organization

	err := tx.Session().First(&organization, "id = ?", organizationID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Organization{}, false, nil
	}
	if err != nil {
		return domain.Organization{}, false, fmt.Errorf("resolve organization: %w", err)
	}

	return organization, true, nil
}

func (r *ResolutionRepository) TreasuryAddress(
	tx database.Tx,
	organizationID uuid.UUID,
) (string, error) {
	if err := tx.Bound(); err != nil {
		return "", err
	}
	if err := grant(tx, resolvingOrganization, organizationID); err != nil {
		return "", fmt.Errorf("grant organization resolution: %w", err)
	}

	var address string

	err := tx.Session().
		Table("identity.treasury_addresses").
		Select("address").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Scan(&address).Error
	if err != nil {
		return "", fmt.Errorf("resolve treasury address: %w", err)
	}

	return address, nil
}

func (r *ResolutionRepository) Facility(
	tx database.Tx,
	facilityID uuid.UUID,
) (domain.Facility, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Facility{}, false, err
	}
	if err := grant(tx, resolvingFacility, facilityID); err != nil {
		return domain.Facility{}, false, fmt.Errorf("grant facility resolution: %w", err)
	}

	var facility domain.Facility

	err := tx.Session().First(&facility, "id = ?", facilityID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Facility{}, false, nil
	}
	if err != nil {
		return domain.Facility{}, false, fmt.Errorf("resolve facility: %w", err)
	}

	return facility, true, nil
}
