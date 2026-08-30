package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
)

type PlanReader interface {
	ListCurrent(ctx context.Context) ([]domain.Plan, error)
}

type PlanRepository struct {
	database *gorm.DB
}

func NewPlanRepository(database *gorm.DB) *PlanRepository {
	return &PlanRepository{database: database}
}

func (r *PlanRepository) ListCurrent(ctx context.Context) ([]domain.Plan, error) {
	var plans []domain.Plan

	err := r.database.WithContext(ctx).
		Preload("Limits").
		Where("effective_to IS NULL").
		Order("monthly_price_usd ASC").
		Find(&plans).Error
	if err != nil {
		return nil, fmt.Errorf("list current plans: %w", err)
	}

	return plans, nil
}
