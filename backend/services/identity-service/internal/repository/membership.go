package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

var ErrNoMembership = errors.New("no active membership")

type OrganizationSnapshot struct {
	Organization       domain.Organization
	TreasuryDesignated bool
}

type MembershipStore interface {
	FindActiveForUser(ctx context.Context, userID uuid.UUID) (domain.OrganizationMembership, error)
	FindOrganization(ctx context.Context, organizationID uuid.UUID) (OrganizationSnapshot, error)
}

type MembershipRepository struct {
	database *gorm.DB
}

func NewMembershipRepository(database *gorm.DB) *MembershipRepository {
	return &MembershipRepository{database: database}
}

func (r *MembershipRepository) FindActiveForUser(
	ctx context.Context,
	userID uuid.UUID,
) (domain.OrganizationMembership, error) {
	var membership domain.OrganizationMembership

	err := database.WithinTenant(
		ctx,
		r.database,
		database.TenantContext{UserID: userID.String()},
		func(tx database.Tx) error {
			return tx.Session().
				Where("user_id = ? AND state = ?", userID, domain.MembershipActive).
				Order("created_at ASC").
				First(&membership).Error
		},
	)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.OrganizationMembership{}, ErrNoMembership
	}
	if err != nil {
		return domain.OrganizationMembership{}, fmt.Errorf("find active membership: %w", err)
	}

	return membership, nil
}

func (r *MembershipRepository) FindOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
) (OrganizationSnapshot, error) {
	var snapshot OrganizationSnapshot

	err := database.WithinTenant(
		ctx,
		r.database,
		database.TenantContext{OrganizationID: organizationID.String()},
		func(tx database.Tx) error {
			if err := tx.Session().First(&snapshot.Organization, "id = ?", organizationID).Error; err != nil {
				return err
			}

			var designated int64
			if err := tx.Session().Model(&domain.TreasuryAddress{}).
				Where("organization_id = ? AND state = ?", organizationID, domain.TreasuryActive).
				Count(&designated).Error; err != nil {
				return err
			}

			snapshot.TreasuryDesignated = designated > 0
			return nil
		},
	)
	if err != nil {
		return OrganizationSnapshot{}, fmt.Errorf("find organization: %w", err)
	}

	return snapshot, nil
}
