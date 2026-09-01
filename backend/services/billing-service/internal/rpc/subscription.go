package rpc

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
)

var subscriptionStates = map[domain.SubscriptionState]billingv1.SubscriptionState{
	domain.SubscriptionActive:      billingv1.SubscriptionState_SUBSCRIPTION_STATE_ACTIVE,
	domain.SubscriptionGracePeriod: billingv1.SubscriptionState_SUBSCRIPTION_STATE_GRACE_PERIOD,
	domain.SubscriptionReadOnly:    billingv1.SubscriptionState_SUBSCRIPTION_STATE_READ_ONLY,
	domain.SubscriptionCancelled:   billingv1.SubscriptionState_SUBSCRIPTION_STATE_CANCELLED,
}

func (s *BillingServer) GetSubscription(
	ctx context.Context,
	request *billingv1.GetSubscriptionRequest,
) (*billingv1.GetSubscriptionResponse, error) {
	organizationID, err := uuid.Parse(request.GetOrganizationId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "organization_id must be a uuid")
	}

	found, err := s.subscriptions.ForOrganization(ctx, organizationID)
	if err != nil {
		s.logger.Error("get subscription failed",
			slog.Any("error", err),
			slog.String("request_id", logging.CorrelationIDFrom(ctx)),
		)
		return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
	}

	if found == nil {
		return &billingv1.GetSubscriptionResponse{}, nil
	}

	return &billingv1.GetSubscriptionResponse{
		Subscription: &billingv1.Subscription{
			OrganizationId: found.OrganizationID.String(),
			PlanTier:       tierToProto[found.Plan.Tier],
			State:          subscriptionStates[found.State],
		},
	}, nil
}
