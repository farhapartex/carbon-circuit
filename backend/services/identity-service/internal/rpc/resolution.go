package rpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type RecordResolver interface {
	Organization(ctx context.Context, organizationID uuid.UUID) (service.ResolvedOrganization, error)
	Facility(ctx context.Context, facilityID uuid.UUID) (service.ResolvedFacility, error)
}

func (s *IdentityServer) ResolveOrganization(
	ctx context.Context,
	request *identityv1.ResolveOrganizationRequest,
) (*identityv1.ResolveOrganizationResponse, error) {
	organizationID, err := uuid.Parse(request.GetOrganizationId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "organization id is not a valid identifier")
	}

	resolved, err := s.resolution.Organization(ctx, organizationID)
	if err != nil {
		if errors.Is(err, service.ErrNotResolvable) {
			return nil, status.Error(codes.NotFound, "no such organization")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identityv1.ResolveOrganizationResponse{
		Id:                     resolved.ID.String(),
		Name:                   resolved.Name,
		CountryOfIncorporation: resolved.CountryOfIncorporation,
		TreasuryAddress:        resolved.TreasuryAddress,
	}, nil
}

func (s *IdentityServer) ResolveFacility(
	ctx context.Context,
	request *identityv1.ResolveFacilityRequest,
) (*identityv1.ResolveFacilityResponse, error) {
	facilityID, err := uuid.Parse(request.GetFacilityId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "facility id is not a valid identifier")
	}

	resolved, err := s.resolution.Facility(ctx, facilityID)
	if err != nil {
		if errors.Is(err, service.ErrNotResolvable) {
			return nil, status.Error(codes.NotFound, "no such facility")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identityv1.ResolveFacilityResponse{
		Id:             resolved.ID.String(),
		OrganizationId: resolved.OrganizationID.String(),
		Name:           resolved.Name,
		CountryCode:    resolved.CountryCode,
	}, nil
}
