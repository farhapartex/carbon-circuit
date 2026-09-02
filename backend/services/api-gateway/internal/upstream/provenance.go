package upstream

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	provenancev1 "github.com/carboncircuit/backend/gen/carboncircuit/provenance/v1"
)

type Provenance struct {
	connection  *grpc.ClientConn
	client      provenancev1.ProvenanceServiceClient
	callTimeout time.Duration
}

func DialProvenance(
	address string,
	callTimeout time.Duration,
	transport credentials.TransportCredentials,
) (*Provenance, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(transport))
	if err != nil {
		return nil, err
	}

	return &Provenance{
		connection:  connection,
		client:      provenancev1.NewProvenanceServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (p *Provenance) Close() error { return p.connection.Close() }

func (p *Provenance) call(
	ctx context.Context,
	idempotencyKey string,
) (context.Context, context.CancelFunc) {
	return callContext(ctx, idempotencyKey, p.callTimeout)
}

func (p *Provenance) CreateBatch(
	ctx context.Context,
	idempotencyKey string,
	request *provenancev1.CreateBatchRequest,
) (*provenancev1.CreateBatchResponse, error) {
	callCtx, cancel := p.call(ctx, idempotencyKey)
	defer cancel()
	return p.client.CreateBatch(callCtx, request)
}

func (p *Provenance) ListBatches(
	ctx context.Context,
	after string,
	limit int32,
) (*provenancev1.ListBatchesResponse, error) {
	callCtx, cancel := p.call(ctx, "")
	defer cancel()
	return p.client.ListBatches(callCtx, &provenancev1.ListBatchesRequest{
		After: after,
		Limit: limit,
	})
}

func (p *Provenance) GetBatch(
	ctx context.Context,
	batchID string,
) (*provenancev1.GetBatchResponse, error) {
	callCtx, cancel := p.call(ctx, "")
	defer cancel()
	return p.client.GetBatch(callCtx, &provenancev1.GetBatchRequest{BatchId: batchID})
}

func (p *Provenance) ListCheckpoints(
	ctx context.Context,
	batchID, after string,
	limit int32,
) (*provenancev1.ListCheckpointsResponse, error) {
	callCtx, cancel := p.call(ctx, "")
	defer cancel()
	return p.client.ListCheckpoints(callCtx, &provenancev1.ListCheckpointsRequest{
		BatchId: batchID,
		After:   after,
		Limit:   limit,
	})
}

func (p *Provenance) LogCheckpoint(
	ctx context.Context,
	idempotencyKey string,
	request *provenancev1.LogCheckpointRequest,
) (*provenancev1.LogCheckpointResponse, error) {
	callCtx, cancel := p.call(ctx, idempotencyKey)
	defer cancel()
	return p.client.LogCheckpoint(callCtx, request)
}

func (p *Provenance) GetComponentBatch(
	ctx context.Context,
	batchID, componentBatchID string,
) (*provenancev1.GetComponentBatchResponse, error) {
	callCtx, cancel := p.call(ctx, "")
	defer cancel()
	return p.client.GetComponentBatch(callCtx, &provenancev1.GetComponentBatchRequest{
		BatchId:          batchID,
		ComponentBatchId: componentBatchID,
	})
}
