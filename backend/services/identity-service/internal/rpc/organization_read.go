package rpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/registry"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type OrganizationDescriber interface {
	Detail(ctx context.Context, organizationID uuid.UUID) (service.Detail, error)
}

var productCategoryToProto = map[string]identityv1.ProductCategory{
	"electronics": identityv1.ProductCategory_PRODUCT_CATEGORY_ELECTRONICS,
	"agriculture": identityv1.ProductCategory_PRODUCT_CATEGORY_AGRICULTURE,
	"pharma":      identityv1.ProductCategory_PRODUCT_CATEGORY_PHARMA,
	"textiles":    identityv1.ProductCategory_PRODUCT_CATEGORY_TEXTILES,
}

func (s *IdentityServer) GetOrganization(
	ctx context.Context,
	_ *identityv1.GetOrganizationRequest,
) (*identityv1.GetOrganizationResponse, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return nil, status.Error(codes.Unauthenticated, "a verified organization is required")
	}

	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	detail, err := s.describer.Detail(ctx, organizationID)
	if errors.Is(err, service.ErrOrganizationNotFound) {
		return nil, status.Error(codes.NotFound, "ORGANIZATION_NOT_FOUND")
	}
	if err != nil {
		s.logger.Error("get organization failed",
			slog.String("organization_id", verified.OrganizationID),
			slog.Any("error", err),
		)
		return nil, status.Error(codes.Internal, "could not read organization")
	}

	organization := detail.Organization

	categories := make([]identityv1.ProductCategory, 0, len(organization.ProductCategories))
	for _, name := range organization.ProductCategories {
		if category, known := productCategoryToProto[name]; known {
			categories = append(categories, category)
		}
	}

	return &identityv1.GetOrganizationResponse{
		Organization: &identityv1.OrganizationDetail{
			Id:                         organization.ID.String(),
			Name:                       organization.Name,
			Type:                       organizationTypes[organization.Type],
			State:                      organizationStates[organization.State],
			VerificationStatus:         verificationStatuses[organization.VerificationStatus],
			CountryOfIncorporation:     organization.CountryOfIncorporation,
			BusinessRegistrationNumber: organization.BusinessRegistrationNumber,
			ProductCategories:          categories,
			TreasuryDesignated:         detail.TreasuryDesignated,
			CreatedAt:                  organization.CreatedAt.UTC().Format(time.RFC3339),
		},
		Role:    organizationRoles[roleFrom(verified.Role)],
		Outcome: outcomeToProto(detail.Outcome),
	}, nil
}

func outcomeToProto(outcome registry.Outcome) *identityv1.VerificationOutcome {
	return &identityv1.VerificationOutcome{
		Status:              verificationStatuses[outcome.Status],
		Rejection:           rejectionProto(outcome.Rejection),
		RegistryMatchFound:  outcome.MatchFound,
		NameSimilarity:      outcome.SimilarityString(),
		RegisteredLegalName: outcome.DisclosableName(),
	}
}

func roleFrom(name string) domain.OrganizationRole {
	return domain.OrganizationRole(name)
}
