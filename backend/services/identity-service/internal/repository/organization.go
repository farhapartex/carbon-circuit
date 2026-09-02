package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

var ErrRegistrationTaken = errors.New("registration number already registered")

type RegistryLookup interface {
	FindRecord(tx database.Tx, countryCode, registrationNumber string) (domain.BusinessRegistryRecord, bool, error)
}

type OrganizationReader interface {
	Find(tx database.Tx, organizationID uuid.UUID) (domain.Organization, bool, error)
	ActiveTreasuryAddress(tx database.Tx, organizationID uuid.UUID) (string, bool, error)
	FindRecordByID(tx database.Tx, recordID uuid.UUID) (domain.BusinessRegistryRecord, bool, error)
}

type OrganizationWriter interface {
	Insert(tx database.Tx, organization *domain.Organization) error
	InsertMembership(tx database.Tx, membership *domain.OrganizationMembership) error
	HasActiveMembership(tx database.Tx, userID uuid.UUID) (bool, error)
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

func (r *OrganizationRepository) HasActiveMembership(
	tx database.Tx,
	userID uuid.UUID,
) (bool, error) {
	if err := tx.Bound(); err != nil {
		return false, err
	}

	var found int64
	err := tx.Session().Model(&domain.OrganizationMembership{}).
		Where("user_id = ? AND state = ?", userID, domain.MembershipActive).
		Count(&found).Error
	if err != nil {
		return false, fmt.Errorf("count memberships: %w", err)
	}

	return found > 0, nil
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

func (r *OrganizationRepository) Find(
	tx database.Tx,
	organizationID uuid.UUID,
) (domain.Organization, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.Organization{}, false, err
	}

	var organization domain.Organization

	err := tx.Session().First(&organization, "id = ?", organizationID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Organization{}, false, nil
	}
	if err != nil {
		return domain.Organization{}, false, fmt.Errorf("find organization: %w", err)
	}

	return organization, true, nil
}

func (r *OrganizationRepository) ActiveTreasuryAddress(
	tx database.Tx,
	organizationID uuid.UUID,
) (string, bool, error) {
	if err := tx.Bound(); err != nil {
		return "", false, err
	}

	var treasury domain.TreasuryAddress

	err := tx.Session().
		Where("organization_id = ? AND state = ?", organizationID, domain.TreasuryActive).
		First(&treasury).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("find treasury address: %w", err)
	}

	return treasury.Address, true, nil
}

func (r *OrganizationRepository) FindRecordByID(
	tx database.Tx,
	recordID uuid.UUID,
) (domain.BusinessRegistryRecord, bool, error) {
	if err := tx.Bound(); err != nil {
		return domain.BusinessRegistryRecord{}, false, err
	}

	var record domain.BusinessRegistryRecord

	err := tx.Session().First(&record, "id = ?", recordID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.BusinessRegistryRecord{}, false, nil
	}
	if err != nil {
		return domain.BusinessRegistryRecord{}, false, fmt.Errorf("find registry record: %w", err)
	}

	return record, true, nil
}
