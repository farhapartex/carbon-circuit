package upstream

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
)

type Billing struct {
	connection  *grpc.ClientConn
	client      billingv1.BillingServiceClient
	callTimeout time.Duration
}

func DialBilling(address string, callTimeout time.Duration) (*Billing, error) {
	connection, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Billing{
		connection:  connection,
		client:      billingv1.NewBillingServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (b *Billing) Close() error { return b.connection.Close() }

func (b *Billing) ListPlans(
	ctx context.Context,
	eligibleFor billingv1.OrganizationType,
) (*billingv1.ListPlansResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, b.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))

	return b.client.ListPlans(callCtx, &billingv1.ListPlansRequest{EligibleFor: eligibleFor})
}

func (b *Billing) GetSubscription(
	ctx context.Context,
	organizationID string,
) (*billingv1.GetSubscriptionResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, b.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))

	return b.client.GetSubscription(callCtx, &billingv1.GetSubscriptionRequest{
		OrganizationId: organizationID,
	})
}
