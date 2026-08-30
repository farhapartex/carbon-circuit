package upstream

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
)

type Identity struct {
	connection  *grpc.ClientConn
	client      identityv1.IdentityServiceClient
	callTimeout time.Duration
}

func DialIdentity(address string, callTimeout time.Duration) (*Identity, error) {
	connection, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Identity{
		connection:  connection,
		client:      identityv1.NewIdentityServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (i *Identity) Close() error {
	return i.connection.Close()
}

func (i *Identity) Ping(ctx context.Context) (*identityv1.PingResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))

	return i.client.Ping(callCtx, &identityv1.PingRequest{})
}
