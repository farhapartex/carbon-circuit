package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
)

var (
	ErrSubscriptionExists = errors.New("organization already has a subscription")
	ErrPlanNotAvailable   = errors.New("no current plan for that tier")
)

type SubscriptionWriter interface {
	FindCurrentPlan(tx database.Tx, tier domain.PlanTier) (domain.Plan, error)
	HasSubscription(tx database.Tx, organizationID uuid.UUID) (bool, error)
	Insert(tx database.Tx, subscription *domain.Subscription) error
}

func (r *SubscriptionRepository) FindCurrentPlan(
	tx database.Tx,
	tier domain.PlanTier,
) (domain.Plan, error) {
	if err := tx.Bound(); err != nil {
		return domain.Plan{}, err
	}

	var plan domain.Plan
	err := tx.Session().
		Where("tier = ? AND effective_to IS NULL", tier).
		First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Plan{}, ErrPlanNotAvailable
	}
	if err != nil {
		return domain.Plan{}, fmt.Errorf("find current plan: %w", err)
	}

	return plan, nil
}

func (r *SubscriptionRepository) HasSubscription(
	tx database.Tx,
	organizationID uuid.UUID,
) (bool, error) {
	if err := tx.Bound(); err != nil {
		return false, err
	}

	var found int64
	err := tx.Session().Model(&domain.Subscription{}).
		Where("organization_id = ?", organizationID).
		Count(&found).Error
	if err != nil {
		return false, fmt.Errorf("count subscriptions: %w", err)
	}

	return found > 0, nil
}

func (r *SubscriptionRepository) Insert(
	tx database.Tx,
	subscription *domain.Subscription,
) error {
	if err := tx.Bound(); err != nil {
		return err
	}

	err := tx.Session().Create(subscription).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrSubscriptionExists
	}
	if err != nil {
		return fmt.Errorf("insert subscription: %w", err)
	}

	return nil
}
