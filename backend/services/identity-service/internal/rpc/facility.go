package rpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type FacilityManager interface {
	Create(
		ctx context.Context,
		actor service.Actor,
		declaration service.FacilityDeclaration,
	) (service.RegisteredFacility, error)
	List(ctx context.Context, actor service.Actor) ([]domain.Facility, error)
	Get(
		ctx context.Context,
		actor service.Actor,
		facilityID uuid.UUID,
	) (domain.Facility, error)
}

var facilityTypesToProto = map[domain.FacilityType]identityv1.FacilityType{
	domain.FacilityRawMaterialProcessing: identityv1.FacilityType_FACILITY_TYPE_RAW_MATERIAL_PROCESSING,
	domain.FacilityComponentFabrication:  identityv1.FacilityType_FACILITY_TYPE_COMPONENT_FABRICATION,
	domain.FacilityAssembly:              identityv1.FacilityType_FACILITY_TYPE_ASSEMBLY,
	domain.FacilityDistribution:          identityv1.FacilityType_FACILITY_TYPE_DISTRIBUTION,
}

var facilityTypesFromProto = map[identityv1.FacilityType]domain.FacilityType{
	identityv1.FacilityType_FACILITY_TYPE_RAW_MATERIAL_PROCESSING: domain.FacilityRawMaterialProcessing,
	identityv1.FacilityType_FACILITY_TYPE_COMPONENT_FABRICATION:   domain.FacilityComponentFabrication,
	identityv1.FacilityType_FACILITY_TYPE_ASSEMBLY:                domain.FacilityAssembly,
	identityv1.FacilityType_FACILITY_TYPE_DISTRIBUTION:            domain.FacilityDistribution,
}

var facilityVerificationsToProto = map[domain.FacilityVerification]identityv1.FacilityVerification{
	domain.FacilityMatched:     identityv1.FacilityVerification_FACILITY_VERIFICATION_FACILITY_MATCHED,
	domain.OrganizationMatched: identityv1.FacilityVerification_FACILITY_VERIFICATION_ORGANIZATION_MATCHED,
	domain.SelfDeclared:        identityv1.FacilityVerification_FACILITY_VERIFICATION_SELF_DECLARED,
}

var trustTiersToProto = map[domain.TrustTier]identityv1.TrustTier{
	domain.TrustNew:      identityv1.TrustTier_TRUST_TIER_NEW,
	domain.TrustVerified: identityv1.TrustTier_TRUST_TIER_VERIFIED,
	domain.TrustTrusted:  identityv1.TrustTier_TRUST_TIER_TRUSTED,
}

var gridRegionsToProto = map[string]identityv1.GridRegion{
	"US-CAISO": identityv1.GridRegion_GRID_REGION_US_CAISO,
	"US-ERCOT": identityv1.GridRegion_GRID_REGION_US_ERCOT,
	"US-PJM":   identityv1.GridRegion_GRID_REGION_US_PJM,
	"US-MISO":  identityv1.GridRegion_GRID_REGION_US_MISO,
	"EU-DE":    identityv1.GridRegion_GRID_REGION_EU_DE,
	"EU-FR":    identityv1.GridRegion_GRID_REGION_EU_FR,
	"EU-PL":    identityv1.GridRegion_GRID_REGION_EU_PL,
	"UK":       identityv1.GridRegion_GRID_REGION_UK,
	"CN-East":  identityv1.GridRegion_GRID_REGION_CN_EAST,
	"CN-South": identityv1.GridRegion_GRID_REGION_CN_SOUTH,
	"IN-North": identityv1.GridRegion_GRID_REGION_IN_NORTH,
	"JP":       identityv1.GridRegion_GRID_REGION_JP,
	"KR":       identityv1.GridRegion_GRID_REGION_KR,
	"TW":       identityv1.GridRegion_GRID_REGION_TW,
	"VN":       identityv1.GridRegion_GRID_REGION_VN,
	"MY":       identityv1.GridRegion_GRID_REGION_MY,
	"SG":       identityv1.GridRegion_GRID_REGION_SG,
	"TH":       identityv1.GridRegion_GRID_REGION_TH,
}

var gridRegionsFromProto = invertGridRegions()

func invertGridRegions() map[identityv1.GridRegion]string {
	inverted := make(map[identityv1.GridRegion]string, len(gridRegionsToProto))
	for name, region := range gridRegionsToProto {
		inverted[region] = name
	}
	return inverted
}

func facilityToProto(facility domain.Facility) *identityv1.Facility {
	return &identityv1.Facility{
		Id:                    facility.ID.String(),
		OrganizationId:        facility.OrganizationID.String(),
		Name:                  facility.Name,
		Address:               facility.Address,
		CountryCode:           facility.CountryCode,
		GridRegion:            gridRegionsToProto[facility.GridRegion],
		Type:                  facilityTypesToProto[facility.Type],
		FacilityReference:     stringOrEmpty(facility.FacilityReference),
		VerificationStatus:    facilityVerificationsToProto[facility.VerificationStatus],
		CeilingDiscountFactor: facility.CeilingDiscountFactor,
		TrustTier:             trustTiersToProto[facility.TrustTier],
		DeclaredCapacity:      facility.DeclaredCapacity,
		DeclaredEnergyKwh:     facility.DeclaredEnergyKwh,
		AttestedCapacity:      stringOrEmpty(facility.AttestedCapacity),
		AttestedEnergyKwh:     stringOrEmpty(facility.AttestedEnergyKwh),
		CreatedAt:             facility.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *IdentityServer) CreateFacility(
	ctx context.Context,
	request *identityv1.CreateFacilityRequest,
) (*identityv1.CreateFacilityResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	idempotencyKey := grpcx.IdempotencyKeyFromIncoming(ctx)
	if idempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	registered, err := s.facilities.Create(ctx, actor, service.FacilityDeclaration{
		Name:              request.GetName(),
		Address:           request.GetAddress(),
		CountryCode:       request.GetCountryCode(),
		GridRegion:        gridRegionsFromProto[request.GetGridRegion()],
		Type:              facilityTypesFromProto[request.GetType()],
		FacilityReference: request.GetFacilityReference(),
		DeclaredCapacity:  request.GetDeclaredCapacity(),
		DeclaredEnergyKwh: request.GetDeclaredEnergyKwh(),
		IdempotencyKey:    idempotencyKey,
		RequestBody:       canonicalFacilityRequest(request),
	})
	if err != nil {
		return nil, facilityFailure(err)
	}

	return &identityv1.CreateFacilityResponse{
		Facility: facilityToProto(registered.Facility),
		Replayed: registered.Replayed,
	}, nil
}

func (s *IdentityServer) ListFacilities(
	ctx context.Context,
	_ *identityv1.ListFacilitiesRequest,
) (*identityv1.ListFacilitiesResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	facilities, err := s.facilities.List(ctx, actor)
	if err != nil {
		return nil, facilityFailure(err)
	}

	listed := make([]*identityv1.Facility, 0, len(facilities))
	for _, facility := range facilities {
		listed = append(listed, facilityToProto(facility))
	}

	return &identityv1.ListFacilitiesResponse{Facilities: listed}, nil
}

func (s *IdentityServer) GetFacility(
	ctx context.Context,
	request *identityv1.GetFacilityRequest,
) (*identityv1.GetFacilityResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	facilityID, err := uuid.Parse(request.GetFacilityId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "facility id is not a valid identifier")
	}

	facility, err := s.facilities.Get(ctx, actor, facilityID)
	if err != nil {
		return nil, facilityFailure(err)
	}

	return &identityv1.GetFacilityResponse{Facility: facilityToProto(facility)}, nil
}

func facilityFailure(err error) error {
	switch {
	case errors.Is(err, service.ErrFacilityNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrNotPermitted):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrRequestInProgress):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, service.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrOrganizationRequired):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func canonicalFacilityRequest(request *identityv1.CreateFacilityRequest) []byte {
	parts := []string{
		strings.TrimSpace(request.GetName()),
		strings.TrimSpace(request.GetAddress()),
		strings.ToUpper(strings.TrimSpace(request.GetCountryCode())),
		request.GetGridRegion().String(),
		request.GetType().String(),
		strings.TrimSpace(request.GetFacilityReference()),
		strings.TrimSpace(request.GetDeclaredCapacity()),
		strings.TrimSpace(request.GetDeclaredEnergyKwh()),
	}
	return []byte(strings.Join(parts, "\x1f"))
}
