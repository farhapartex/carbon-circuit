package upstream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/provenance-service/internal/service"
)

type Identity struct {
	connection  *grpc.ClientConn
	client      identityv1.IdentityServiceClient
	callTimeout time.Duration
}

func DialIdentity(
	address string,
	callTimeout time.Duration,
	transport credentials.TransportCredentials,
) (*Identity, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(transport))
	if err != nil {
		return nil, err
	}

	return &Identity{
		connection:  connection,
		client:      identityv1.NewIdentityServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (i *Identity) Close() error { return i.connection.Close() }

func (i *Identity) Facility(
	ctx context.Context,
	facilityID uuid.UUID,
) (service.Facility, error) {
	callCtx, cancel := context.WithTimeout(
		grpcx.ForwardServiceToken(ctx), i.callTimeout,
	)
	defer cancel()

	response, err := i.client.GetFacility(callCtx, &identityv1.GetFacilityRequest{
		FacilityId: facilityID.String(),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return service.Facility{}, service.ErrFacilityUnknown
		}
		return service.Facility{}, fmt.Errorf("resolve facility: %w", err)
	}

	facility := response.GetFacility()
	if facility == nil {
		return service.Facility{}, service.ErrFacilityUnknown
	}

	resolved, err := uuid.Parse(facility.GetId())
	if err != nil {
		return service.Facility{}, errors.New("identity returned an unusable facility id")
	}

	return service.Facility{ID: resolved, Name: facility.GetName()}, nil
}
