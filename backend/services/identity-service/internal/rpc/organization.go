package rpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type OrganizationCreator interface {
	Create(ctx context.Context, subject string, registration service.Registration) (service.Registered, error)
}

var productCategoryNames = map[identityv1.ProductCategory]string{
	identityv1.ProductCategory_PRODUCT_CATEGORY_ELECTRONICS: "electronics",
	identityv1.ProductCategory_PRODUCT_CATEGORY_AGRICULTURE: "agriculture",
	identityv1.ProductCategory_PRODUCT_CATEGORY_PHARMA:      "pharma",
	identityv1.ProductCategory_PRODUCT_CATEGORY_TEXTILES:    "textiles",
}

var rejectionToProto = map[domain.RegistryRejection]identityv1.RegistryRejection{
	domain.RejectionEntityDissolved: identityv1.RegistryRejection_REGISTRY_REJECTION_ENTITY_DISSOLVED,
	domain.RejectionSanctionsFlag:   identityv1.RegistryRejection_REGISTRY_REJECTION_SANCTIONS_FLAG,
	domain.RejectionNameMismatch:    identityv1.RegistryRejection_REGISTRY_REJECTION_NAME_MISMATCH,
}

var organizationTypeFromProto = map[identityv1.OrganizationType]domain.OrganizationType{
	identityv1.OrganizationType_ORGANIZATION_TYPE_MANUFACTURER: domain.OrganizationManufacturer,
	identityv1.OrganizationType_ORGANIZATION_TYPE_ASSEMBLER:    domain.OrganizationAssembler,
	identityv1.OrganizationType_ORGANIZATION_TYPE_LOGISTICS:    domain.OrganizationLogistics,
	identityv1.OrganizationType_ORGANIZATION_TYPE_CREDIT_BUYER: domain.OrganizationCreditBuyer,
}

func (s *IdentityServer) CreateOrganization(
	ctx context.Context,
	request *identityv1.CreateOrganizationRequest,
) (*identityv1.CreateOrganizationResponse, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || verified.Subject == "" {
		return nil, status.Error(codes.Unauthenticated, "a verified caller is required")
	}

	if verified.HasOrganization() {
		return nil, status.Error(codes.AlreadyExists, "ORGANIZATION_EXISTS")
	}

	subject := verified.Subject

	organizationType, known := organizationTypeFromProto[request.GetType()]
	if !known {
		return nil, status.Error(codes.InvalidArgument, "type must be a known organization type")
	}

	idempotencyKey := grpcx.IdempotencyKeyFromIncoming(ctx)
	if idempotencyKey == "" {
		return nil, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	registration := service.Registration{
		LegalName:          strings.TrimSpace(request.GetName()),
		Type:               organizationType,
		CountryCode:        strings.ToUpper(strings.TrimSpace(request.GetCountryOfIncorporation())),
		RegistrationNumber: strings.TrimSpace(request.GetBusinessRegistrationNumber()),
		ProductCategories:  categoriesFrom(request.GetProductCategories()),
		IdempotencyKey:     idempotencyKey,
		RequestBody:        canonicalRequest(request),
	}

	registered, err := s.organizations.Create(ctx, subject, registration)
	if err != nil {
		return nil, createFailure(s.logger, subject, err)
	}

	return &identityv1.CreateOrganizationResponse{
		Organization: organizationToProto(&registered.Organization, false),
		Role:         organizationRoles[registered.Role],
		Outcome: &identityv1.VerificationOutcome{
			Status:             verificationStatuses[registered.Outcome.Status],
			Rejection:          rejectionProto(registered.Outcome.Rejection),
			RegistryMatchFound: registered.Outcome.MatchFound,
			NameSimilarity:     registered.Outcome.SimilarityString(),
		},
	}, nil
}

func categoriesFrom(categories []identityv1.ProductCategory) []string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		if name, known := productCategoryNames[category]; known {
			names = append(names, name)
		}
	}
	return names
}

func rejectionProto(rejection *domain.RegistryRejection) identityv1.RegistryRejection {
	if rejection == nil {
		return identityv1.RegistryRejection_REGISTRY_REJECTION_UNSPECIFIED
	}
	return rejectionToProto[*rejection]
}

func canonicalRequest(request *identityv1.CreateOrganizationRequest) []byte {
	parts := []string{
		strings.TrimSpace(request.GetName()),
		request.GetType().String(),
		strings.ToUpper(strings.TrimSpace(request.GetCountryOfIncorporation())),
		strings.TrimSpace(request.GetBusinessRegistrationNumber()),
		strings.Join(categoriesFrom(request.GetProductCategories()), ","),
	}
	return []byte(strings.Join(parts, "\x1f"))
}

func createFailure(logger *slog.Logger, subject string, err error) error {
	switch {
	case errors.Is(err, service.ErrRequestInProgress):
		return status.Error(codes.Aborted, "REQUEST_IN_PROGRESS")
	case errors.Is(err, service.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, "IDEMPOTENCY_KEY_REUSED")
	case errors.Is(err, service.ErrOrganizationExists):
		return status.Error(codes.AlreadyExists, "ORGANIZATION_EXISTS")
	case errors.Is(err, repository.ErrRegistrationTaken):
		return status.Error(codes.AlreadyExists, "REGISTRATION_TAKEN")
	case errors.Is(err, repository.ErrUserNotFound):
		return status.Error(codes.NotFound, "USER_NOT_FOUND")
	}

	logger.Error("create organization failed",
		slog.String("subject", subject),
		slog.Any("error", err),
	)
	return status.Error(codes.Internal, "could not create organization")
}
