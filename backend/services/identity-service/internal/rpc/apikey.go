package rpc

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type APIKeyManager interface {
	Create(ctx context.Context, actor service.Actor, name string) (service.IssuedAPIKey, error)
	Validate(ctx context.Context, presented string) (service.ValidatedAPIKey, error)
	List(ctx context.Context, actor service.Actor) ([]domain.APIKey, error)
	Revoke(ctx context.Context, actor service.Actor, keyID uuid.UUID) error
}

func apiKeyToProto(key domain.APIKey) *identityv1.APIKey {
	mapped := &identityv1.APIKey{
		Id:        key.ID.String(),
		Name:      key.Name,
		Prefix:    key.Prefix,
		CreatedAt: key.CreatedAt.UTC().Format(time.RFC3339),
	}

	if key.LastUsedAt != nil {
		mapped.LastUsedAt = key.LastUsedAt.UTC().Format(time.RFC3339)
	}
	if key.RevokedAt != nil {
		mapped.RevokedAt = key.RevokedAt.UTC().Format(time.RFC3339)
	}

	return mapped
}

func (s *IdentityServer) CreateAPIKey(
	ctx context.Context,
	request *identityv1.CreateAPIKeyRequest,
) (*identityv1.CreateAPIKeyResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	issued, err := s.apiKeys.Create(ctx, actor, request.GetName())
	if err != nil {
		return nil, apiKeyFailure(err)
	}

	return &identityv1.CreateAPIKeyResponse{
		Key:          apiKeyToProto(issued.Key),
		PresentedKey: issued.Presented,
	}, nil
}

func (s *IdentityServer) ListAPIKeys(
	ctx context.Context,
	_ *identityv1.ListAPIKeysRequest,
) (*identityv1.ListAPIKeysResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	listed, err := s.apiKeys.List(ctx, actor)
	if err != nil {
		return nil, apiKeyFailure(err)
	}

	keys := make([]*identityv1.APIKey, 0, len(listed))
	for _, key := range listed {
		keys = append(keys, apiKeyToProto(key))
	}

	return &identityv1.ListAPIKeysResponse{Keys: keys}, nil
}

func (s *IdentityServer) RevokeAPIKey(
	ctx context.Context,
	request *identityv1.RevokeAPIKeyRequest,
) (*identityv1.RevokeAPIKeyResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	keyID, err := uuid.Parse(request.GetKeyId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "key id is not a valid identifier")
	}

	if err := s.apiKeys.Revoke(ctx, actor, keyID); err != nil {
		return nil, apiKeyFailure(err)
	}

	return &identityv1.RevokeAPIKeyResponse{}, nil
}

func apiKeyFailure(err error) error {
	switch {
	case errors.Is(err, service.ErrAPIKeyNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrNotPermitted):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrAPIKeyNameEmpty):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *IdentityServer) ValidateAPIKey(
	ctx context.Context,
	request *identityv1.ValidateAPIKeyRequest,
) (*identityv1.ValidateAPIKeyResponse, error) {
	validated, err := s.apiKeys.Validate(ctx, request.GetPresentedKey())
	if err != nil {
		if errors.Is(err, service.ErrAPIKeyRejected) ||
			errors.Is(err, service.ErrAPIKeyRevoked) {
			return nil, status.Error(codes.Unauthenticated, "api key is not valid")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identityv1.ValidateAPIKeyResponse{
		KeyId:              validated.KeyID.String(),
		Prefix:             validated.Prefix,
		OrganizationId:     validated.OrganizationID.String(),
		OrganizationName:   validated.OrganizationName,
		OrganizationType:   organizationTypes[validated.OrganizationType],
		OrganizationState:  organizationStates[validated.OrganizationState],
		VerificationStatus: verificationStatuses[validated.VerificationStatus],
		ActingUserId:       validated.ActingUserID.String(),
	}, nil
}
