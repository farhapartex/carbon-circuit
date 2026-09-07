package upstream

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	provenancereadv1 "github.com/carboncircuit/backend/gen/carboncircuit/provenanceread/v1"
)

type ProvenanceRead struct {
	connection  *grpc.ClientConn
	client      provenancereadv1.ProvenanceReadServiceClient
	callTimeout time.Duration
}

func DialProvenanceRead(
	address string,
	callTimeout time.Duration,
	transport credentials.TransportCredentials,
) (*ProvenanceRead, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(transport))
	if err != nil {
		return nil, err
	}

	return &ProvenanceRead{
		connection:  connection,
		client:      provenancereadv1.NewProvenanceReadServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (p *ProvenanceRead) Close() error { return p.connection.Close() }

func (p *ProvenanceRead) TrackBatch(
	ctx context.Context,
	reference string,
) (*provenancereadv1.TrackBatchResponse, error) {
	callCtx, cancel := callContext(ctx, "", p.callTimeout)
	defer cancel()
	return p.client.TrackBatch(callCtx, &provenancereadv1.TrackBatchRequest{
		PublicReference: reference,
	})
}

func (p *ProvenanceRead) Ping(ctx context.Context) error {
	callCtx, cancel := callContext(ctx, "", p.callTimeout)
	defer cancel()
	_, err := p.client.Ping(callCtx, &provenancereadv1.PingRequest{})
	return err
}
