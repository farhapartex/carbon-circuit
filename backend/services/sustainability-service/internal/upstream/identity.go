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
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

var gridRegionName = map[identityv1.GridRegion]string{
	identityv1.GridRegion_GRID_REGION_US_CAISO: "US-CAISO",
	identityv1.GridRegion_GRID_REGION_US_ERCOT: "US-ERCOT",
	identityv1.GridRegion_GRID_REGION_US_PJM:   "US-PJM",
	identityv1.GridRegion_GRID_REGION_US_MISO:  "US-MISO",
	identityv1.GridRegion_GRID_REGION_EU_DE:    "EU-DE",
	identityv1.GridRegion_GRID_REGION_EU_FR:    "EU-FR",
	identityv1.GridRegion_GRID_REGION_EU_PL:    "EU-PL",
	identityv1.GridRegion_GRID_REGION_UK:       "UK",
	identityv1.GridRegion_GRID_REGION_CN_EAST:  "CN-East",
	identityv1.GridRegion_GRID_REGION_CN_SOUTH: "CN-South",
	identityv1.GridRegion_GRID_REGION_IN_NORTH: "IN-North",
	identityv1.GridRegion_GRID_REGION_JP:       "JP",
	identityv1.GridRegion_GRID_REGION_KR:       "KR",
	identityv1.GridRegion_GRID_REGION_TW:       "TW",
	identityv1.GridRegion_GRID_REGION_VN:       "VN",
	identityv1.GridRegion_GRID_REGION_MY:       "MY",
	identityv1.GridRegion_GRID_REGION_SG:       "SG",
	identityv1.GridRegion_GRID_REGION_TH:       "TH",
}

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
		return service.Facility{}, fmt.Errorf("identity returned an unusable facility id")
	}

	return service.Facility{
		ID:                    resolved,
		Name:                  facility.GetName(),
		GridRegion:            gridRegionName[facility.GetGridRegion()],
		CeilingDiscountFactor: facility.GetCeilingDiscountFactor(),
		DeclaredEnergyKwh:     facility.GetDeclaredEnergyKwh(),
		AttestedEnergyKwh:     facility.GetAttestedEnergyKwh(),
	}, nil
}

func (i *Identity) Ping(ctx context.Context) error {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	_, err := i.client.Ping(callCtx, &identityv1.PingRequest{})
	return err
}
