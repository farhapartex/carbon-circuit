package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
)

var ErrNoSubscription = errors.New("no subscription for organization")

type SubscriptionReader interface {
	FindForOrganization(ctx context.Context, organizationID uuid.UUID) (domain.Subscription, error)
}

type SubscriptionRepository struct {
	database *gorm.DB
}

func NewSubscriptionRepository(database *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{database: database}
}

func (r *SubscriptionRepository) FindForOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
) (domain.Subscription, error) {
	var subscription domain.Subscription

	err := database.WithinTenant(
		ctx,
		r.database,
		database.TenantContext{OrganizationID: organizationID.String()},
		func(tx database.Tx) error {
			return tx.Session().
				Preload("Plan").
				Where("organization_id = ?", organizationID).
				First(&subscription).Error
		},
	)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Subscription{}, ErrNoSubscription
	}
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("find subscription: %w", err)
	}

	return subscription, nil
}
