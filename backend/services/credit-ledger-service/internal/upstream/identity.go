package upstream

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/service"
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

func (i *Identity) call(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, i.callTimeout)
}

func (i *Identity) Facility(
	ctx context.Context,
	facilityID uuid.UUID,
) (service.Facility, error) {
	callCtx, cancel := i.call(ctx)
	defer cancel()

	resolved, err := i.client.ResolveFacility(callCtx, &identityv1.ResolveFacilityRequest{
		FacilityId: facilityID.String(),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return service.Facility{}, fmt.Errorf("facility %s is not registered", facilityID)
		}
		return service.Facility{}, fmt.Errorf("resolve facility: %w", err)
	}

	return service.Facility{
		ID:          facilityID,
		Name:        resolved.GetName(),
		CountryCode: resolved.GetCountryCode(),
	}, nil
}

func (i *Identity) Treasury(
	ctx context.Context,
	organizationID uuid.UUID,
) (service.Treasury, error) {
	callCtx, cancel := i.call(ctx)
	defer cancel()

	resolved, err := i.client.ResolveOrganization(callCtx, &identityv1.ResolveOrganizationRequest{
		OrganizationId: organizationID.String(),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return service.Treasury{}, fmt.Errorf("organization %s is not registered", organizationID)
		}
		return service.Treasury{}, fmt.Errorf("resolve organization: %w", err)
	}

	return service.Treasury{Address: resolved.GetTreasuryAddress()}, nil
}

func (i *Identity) Ping(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	_, err := i.client.Ping(callCtx, &identityv1.PingRequest{})
	return err
}
