package rpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
	"github.com/carboncircuit/backend/services/billing-service/internal/service"
)

var organizationTypeByName = map[string]domain.OrganizationType{
	"manufacturer": domain.OrganizationManufacturer,
	"assembler":    domain.OrganizationAssembler,
	"logistics":    domain.OrganizationLogistics,
	"credit_buyer": domain.OrganizationCreditBuyer,
}

var tierFromProto = map[billingv1.PlanTier]domain.PlanTier{
	billingv1.PlanTier_PLAN_TIER_BUYER:      domain.TierBuyer,
	billingv1.PlanTier_PLAN_TIER_STARTER:    domain.TierStarter,
	billingv1.PlanTier_PLAN_TIER_GROWTH:     domain.TierGrowth,
	billingv1.PlanTier_PLAN_TIER_ENTERPRISE: domain.TierEnterprise,
}

func (s *BillingServer) CreateSubscription(
	ctx context.Context,
	request *billingv1.CreateSubscriptionRequest,
) (*billingv1.CreateSubscriptionResponse, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return nil, status.Error(codes.Unauthenticated, "a verified organization is required")
	}

	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	tier, known := tierFromProto[request.GetPlanTier()]
	if !known {
		return nil, status.Error(codes.InvalidArgument, "plan_tier must be a known tier")
	}

	organizationType, known := organizationTypeByName[verified.OrganizationType]
	if !known {
		return nil, status.Error(codes.Unauthenticated, "service token carries an unknown organization type")
	}

	key := grpcx.IdempotencyKeyFromIncoming(ctx)
	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	enrolled, err := s.creator.Create(ctx, service.Enrolment{
		OrganizationID:   organizationID,
		OrganizationType: organizationType,
		Tier:             tier,
		IdempotencyKey:   key,
		RequestBody:      []byte(organizationID.String() + "\x1f" + string(tier)),
	})
	if err != nil {
		return nil, s.enrolmentFailure(ctx, organizationID, err)
	}

	subscription := enrolled.Subscription

	return &billingv1.CreateSubscriptionResponse{
		Subscription: &billingv1.Subscription{
			OrganizationId: subscription.OrganizationID.String(),
			PlanTier:       tierToProto[subscription.Plan.Tier],
			State:          subscriptionStates[subscription.State],
		},
	}, nil
}

func (s *BillingServer) enrolmentFailure(
	ctx context.Context,
	organizationID uuid.UUID,
	err error,
) error {
	switch {
	case errors.Is(err, service.ErrRequestInProgress):
		return status.Error(codes.Aborted, "REQUEST_IN_PROGRESS")
	case errors.Is(err, service.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, "IDEMPOTENCY_KEY_REUSED")
	case errors.Is(err, service.ErrSubscriptionExists):
		return status.Error(codes.AlreadyExists, "SUBSCRIPTION_EXISTS")
	case errors.Is(err, service.ErrPlanNotAllowed):
		return status.Error(codes.PermissionDenied, "PLAN_NOT_ALLOWED")
	}

	s.logger.Error("create subscription failed",
		slog.String("organization_id", organizationID.String()),
		slog.String("request_id", logging.CorrelationIDFrom(ctx)),
		slog.Any("error", err),
	)
	return status.Error(codes.Internal, "could not create subscription")
}
