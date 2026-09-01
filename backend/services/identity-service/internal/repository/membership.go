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

type MembershipStore interface {
	FindActiveForUser(ctx context.Context, userID uuid.UUID) (domain.OrganizationMembership, error)
	FindOrganization(ctx context.Context, organizationID uuid.UUID) (domain.Organization, error)
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
		func(tx *gorm.DB) error {
			return tx.Where("user_id = ? AND state = ?", userID, domain.MembershipActive).
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
) (domain.Organization, error) {
	var organization domain.Organization

	err := database.WithinTenant(
		ctx,
		r.database,
		database.TenantContext{OrganizationID: organizationID.String()},
		func(tx *gorm.DB) error {
			return tx.First(&organization, "id = ?", organizationID).Error
		},
	)
	if err != nil {
		return domain.Organization{}, fmt.Errorf("find organization: %w", err)
	}

	return organization, nil
}
