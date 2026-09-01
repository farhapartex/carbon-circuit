package rpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type SessionResolver interface {
	Resolve(ctx context.Context, claims service.Claims) (service.Session, error)
}

func (s *IdentityServer) ResolveSession(
	ctx context.Context,
	request *identityv1.ResolveSessionRequest,
) (*identityv1.ResolveSessionResponse, error) {
	claims, err := claimsFrom(request)
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.Resolve(ctx, claims)
	if errors.Is(err, service.ErrUnverifiedEmailClaim) {
		return nil, status.Error(codes.FailedPrecondition, "email must be verified before it can be linked")
	}
	if err != nil {
		s.logger.Error("resolve session failed",
			slog.String("subject", request.GetAuth0Subject()),
			slog.Any("error", err),
		)
		return nil, status.Error(codes.Internal, "could not resolve session")
	}

	return &identityv1.ResolveSessionResponse{
		User:            userToProto(session.User),
		NeedsOnboarding: session.NeedsOnboarding(),
		Organization:    organizationToProto(session.Organization),
		Role:            organizationRoles[session.Role],
	}, nil
}

func claimsFrom(request *identityv1.ResolveSessionRequest) (service.Claims, error) {
	subject := strings.TrimSpace(request.GetAuth0Subject())
	if subject == "" {
		return service.Claims{}, status.Error(codes.InvalidArgument, "auth0_subject is required")
	}

	email := strings.TrimSpace(request.GetEmail())
	if email == "" {
		return service.Claims{}, status.Error(codes.InvalidArgument, "email is required")
	}

	name := strings.TrimSpace(request.GetName())
	if name == "" {
		name = email
	}

	return service.Claims{
		Auth0Subject:  subject,
		Email:         email,
		EmailVerified: request.GetEmailVerified(),
		Name:          name,
	}, nil
}
