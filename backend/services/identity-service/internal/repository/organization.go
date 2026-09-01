package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

var ErrRegistrationTaken = errors.New("registration number already registered")

type RegistryLookup interface {
	FindRecord(tx database.Tx, countryCode, registrationNumber string) (domain.BusinessRegistryRecord, bool, error)
}

type OrganizationWriter interface {
	Insert(tx database.Tx, organization *domain.Organization) error
	InsertMembership(tx database.Tx, membership *domain.OrganizationMembership) error
}

type OrganizationRepository struct{}

func NewOrganizationRepository() *OrganizationRepository {
	return &OrganizationRepository{}
}

func (r *OrganizationRepository) FindRecord(
	tx database.Tx,
	countryCode, registrationNumber string,
) (domain.BusinessRegistryRecord, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.BusinessRegistryRecord{}, false, err
	}

	var record domain.BusinessRegistryRecord

	err := tx.Session().
		Where("country_code = ? AND registration_number = ?", countryCode, registrationNumber).
		First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.BusinessRegistryRecord{}, false, nil
	}
	if err != nil {
		return domain.BusinessRegistryRecord{}, false, fmt.Errorf("lookup registry record: %w", err)
	}

	return record, true, nil
}

func (r *OrganizationRepository) Insert(tx database.Tx, organization *domain.Organization) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(organization).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrRegistrationTaken
	}
	if err != nil {
		return fmt.Errorf("insert organization: %w", err)
	}

	return nil
}

func (r *OrganizationRepository) InsertMembership(
	tx database.Tx,
	membership *domain.OrganizationMembership,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	if err := tx.Session().Create(membership).Error; err != nil {
		return fmt.Errorf("insert membership: %w", err)
	}

	return nil
}
