package upstream

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sustainabilityv1 "github.com/carboncircuit/backend/gen/carboncircuit/sustainability/v1"
)

type Sustainability struct {
	connection  *grpc.ClientConn
	client      sustainabilityv1.SustainabilityServiceClient
	callTimeout time.Duration
}

func DialSustainability(
	address string,
	callTimeout time.Duration,
	transport credentials.TransportCredentials,
) (*Sustainability, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(transport))
	if err != nil {
		return nil, err
	}

	return &Sustainability{
		connection:  connection,
		client:      sustainabilityv1.NewSustainabilityServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (s *Sustainability) Close() error { return s.connection.Close() }

func (s *Sustainability) SubmitClaim(
	ctx context.Context,
	idempotencyKey string,
	request *sustainabilityv1.SubmitClaimRequest,
) (*sustainabilityv1.SubmitClaimResponse, error) {
	callCtx, cancel := callContext(ctx, idempotencyKey, s.callTimeout)
	defer cancel()
	return s.client.SubmitClaim(callCtx, request)
}

func (s *Sustainability) ListClaims(
	ctx context.Context,
	request *sustainabilityv1.ListClaimsRequest,
) (*sustainabilityv1.ListClaimsResponse, error) {
	callCtx, cancel := callContext(ctx, "", s.callTimeout)
	defer cancel()
	return s.client.ListClaims(callCtx, request)
}

func (s *Sustainability) GetClaim(
	ctx context.Context,
	claimID string,
) (*sustainabilityv1.GetClaimResponse, error) {
	callCtx, cancel := callContext(ctx, "", s.callTimeout)
	defer cancel()
	return s.client.GetClaim(callCtx, &sustainabilityv1.GetClaimRequest{ClaimId: claimID})
}

func (s *Sustainability) PreviewCeiling(
	ctx context.Context,
	request *sustainabilityv1.PreviewCeilingRequest,
) (*sustainabilityv1.PreviewCeilingResponse, error) {
	callCtx, cancel := callContext(ctx, "", s.callTimeout)
	defer cancel()
	return s.client.PreviewCeiling(callCtx, request)
}

func (s *Sustainability) Ping(ctx context.Context) (*sustainabilityv1.PingResponse, error) {
	callCtx, cancel := callContext(ctx, "", s.callTimeout)
	defer cancel()
	return s.client.Ping(callCtx, &sustainabilityv1.PingRequest{})
}

func (s *Sustainability) ReviewQueue(
	ctx context.Context,
	request *sustainabilityv1.ReviewQueueRequest,
) (*sustainabilityv1.ReviewQueueResponse, error) {
	callCtx, cancel := callContext(ctx, "", s.callTimeout)
	defer cancel()
	return s.client.ReviewQueue(callCtx, request)
}

func (s *Sustainability) ReviewClaim(
	ctx context.Context,
	claimID string,
) (*sustainabilityv1.ReviewClaimResponse, error) {
	callCtx, cancel := callContext(ctx, "", s.callTimeout)
	defer cancel()
	return s.client.ReviewClaim(callCtx, &sustainabilityv1.ReviewClaimRequest{ClaimId: claimID})
}

func (s *Sustainability) DecideClaim(
	ctx context.Context,
	idempotencyKey string,
	request *sustainabilityv1.DecideClaimRequest,
) (*sustainabilityv1.DecideClaimResponse, error) {
	callCtx, cancel := callContext(ctx, idempotencyKey, s.callTimeout)
	defer cancel()
	return s.client.DecideClaim(callCtx, request)
}
