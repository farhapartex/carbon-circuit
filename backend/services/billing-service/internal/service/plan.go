package service

import (
	"context"
	"time"

	"github.com/carboncircuit/backend/internal/cache"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
	"github.com/carboncircuit/backend/services/billing-service/internal/repository"
)

const (
	PlanCacheKey = "billing:plans:v1"
	PlanCacheTTL = 24 * time.Hour
)

type PlanService struct {
	plans repository.PlanReader
	cache *cache.Client
}

func NewPlanService(plans repository.PlanReader, cacheClient *cache.Client) *PlanService {
	return &PlanService{plans: plans, cache: cacheClient}
}

func (s *PlanService) List(
	ctx context.Context,
	eligibleFor domain.OrganizationType,
) ([]domain.Plan, error) {
	current, err := cache.ReadThrough(ctx, s.cache, PlanCacheKey, PlanCacheTTL, s.plans.ListCurrent)
	if err != nil {
		return nil, err
	}

	if eligibleFor == "" {
		return current, nil
	}

	eligible := make([]domain.Plan, 0, len(current))
	for _, plan := range current {
		if plan.AllowsOrganizationType(eligibleFor) {
			eligible = append(eligible, plan)
		}
	}

	return eligible, nil
}

func (s *PlanService) InvalidateCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	s.cache.Invalidate(ctx, PlanCacheKey)
}
