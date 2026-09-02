package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

type FacilityStore interface {
	Insert(tx database.Tx, facility *domain.Facility) error
	List(tx database.Tx, organizationID uuid.UUID) ([]domain.Facility, error)
	Find(tx database.Tx, organizationID, facilityID uuid.UUID) (domain.Facility, bool, error)
	FindRegistryRecord(
		tx database.Tx,
		registrationNumber, facilityReference string,
	) (domain.FacilityRegistryRecord, bool, error)
}

type FacilityRepository struct{}

func NewFacilityRepository() *FacilityRepository { return &FacilityRepository{} }

func (r *FacilityRepository) Insert(tx database.Tx, facility *domain.Facility) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	if err := tx.Session().Create(facility).Error; err != nil {
		return fmt.Errorf("insert facility: %w", err)
	}

	return nil
}

func (r *FacilityRepository) List(
	tx database.Tx,
	organizationID uuid.UUID,
) ([]domain.Facility, error) {
	if err := tx.Bound(); err != nil {
		return nil, err
	}

	var facilities []domain.Facility

	err := tx.Session().
		Where("organization_id = ?", organizationID).
		Order("created_at DESC").
		Find(&facilities).Error
	if err != nil {
		return nil, fmt.Errorf("list facilities: %w", err)
	}

	return facilities, nil
}

func (r *FacilityRepository) Find(
	tx database.Tx,
	organizationID, facilityID uuid.UUID,
) (domain.Facility, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Facility{}, false, err
	}

	var facility domain.Facility

	err := tx.Session().
		Where("id = ? AND organization_id = ?", facilityID, organizationID).
		First(&facility).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Facility{}, false, nil
	}
	if err != nil {
		return domain.Facility{}, false, fmt.Errorf("find facility: %w", err)
	}

	return facility, true, nil
}

func (r *FacilityRepository) FindRegistryRecord(
	tx database.Tx,
	registrationNumber, facilityReference string,
) (domain.FacilityRegistryRecord, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.FacilityRegistryRecord{}, false, err
	}

	var record domain.FacilityRegistryRecord

	err := tx.Session().
		Where(
			"organization_registration_number = ? AND facility_reference = ?",
			registrationNumber, facilityReference,
		).
		First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.FacilityRegistryRecord{}, false, nil
	}
	if err != nil {
		return domain.FacilityRegistryRecord{}, false, fmt.Errorf("lookup facility registry: %w", err)
	}

	return record, true, nil
}
