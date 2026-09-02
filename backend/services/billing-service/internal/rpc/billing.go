package rpc

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/services/billing-service/internal/service"
)

type BillingServer struct {
	billingv1.UnimplementedBillingServiceServer

	plans         *service.PlanService
	subscriptions *service.SubscriptionService
	creator       *service.SubscriptionCreator
	logger        *slog.Logger
}

func NewBillingServer(
	plans *service.PlanService,
	subscriptions *service.SubscriptionService,
	creator *service.SubscriptionCreator,
	logger *slog.Logger,
) *BillingServer {
	return &BillingServer{
		plans:         plans,
		subscriptions: subscriptions,
		creator:       creator,
		logger:        logger,
	}
}

func (s *BillingServer) ListPlans(
	ctx context.Context,
	request *billingv1.ListPlansRequest,
) (*billingv1.ListPlansResponse, error) {
	eligibleFor := organizationTypeFromProto[request.GetEligibleFor()]

	found, err := s.plans.List(ctx, eligibleFor)
	if err != nil {
		s.logger.Error("list plans failed",
			slog.Any("error", err),
			slog.String("request_id", logging.CorrelationIDFrom(ctx)),
		)
		return nil, status.Error(codes.Internal, "INTERNAL_ERROR")
	}

	response := &billingv1.ListPlansResponse{Plans: make([]*billingv1.Plan, 0, len(found))}
	for _, plan := range found {
		response.Plans = append(response.Plans, planToProto(plan))
	}

	return response, nil
}
