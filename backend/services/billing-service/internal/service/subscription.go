package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
	"github.com/carboncircuit/backend/services/billing-service/internal/repository"
)

type SubscriptionService struct {
	subscriptions repository.SubscriptionReader
}

func NewSubscriptionService(subscriptions repository.SubscriptionReader) *SubscriptionService {
	return &SubscriptionService{subscriptions: subscriptions}
}

func (s *SubscriptionService) ForOrganization(
	ctx context.Context,
	organizationID uuid.UUID,
) (*domain.Subscription, error) {
	found, err := s.subscriptions.FindForOrganization(ctx, organizationID)
	if errors.Is(err, repository.ErrNoSubscription) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &found, nil
}
