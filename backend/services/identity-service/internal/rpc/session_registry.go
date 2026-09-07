package rpc

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type SessionRegistrar interface {
	Record(ctx context.Context, sighting service.SessionSighting) error
	List(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	Revoke(ctx context.Context, userID uuid.UUID, auth0SessionID string) error
}

func (s *IdentityServer) signedInUser(
	ctx context.Context,
) (uuid.UUID, string, string, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || verified.UserID == "" {
		return uuid.Nil, "", "", status.Error(codes.Unauthenticated, "a verified user is required")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return uuid.Nil, "", "", status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	return userID, verified.SessionID, verified.Subject, nil
}

func (s *IdentityServer) RecordSession(
	ctx context.Context,
	request *identityv1.RecordSessionRequest,
) (*identityv1.RecordSessionResponse, error) {
	userID, sessionID, _, err := s.signedInUser(ctx)
	if err != nil {
		return nil, err
	}

	err = s.registry.Record(ctx, service.SessionSighting{
		UserID:         userID,
		Auth0SessionID: sessionID,
		UserAgent:      request.GetUserAgent(),
		IPAddress:      request.GetIpAddress(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identityv1.RecordSessionResponse{}, nil
}

func (s *IdentityServer) ListSessions(
	ctx context.Context,
	_ *identityv1.ListSessionsRequest,
) (*identityv1.ListSessionsResponse, error) {
	userID, _, _, err := s.signedInUser(ctx)
	if err != nil {
		return nil, err
	}

	listed, err := s.registry.List(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sessions := make([]*identityv1.Session, 0, len(listed))
	for _, session := range listed {
		sessions = append(sessions, &identityv1.Session{
			Auth0SessionId: session.Auth0SessionID,
			UserAgent:      session.UserAgent,
			IpAddress:      session.IPAddress,
			StartedAt:      session.StartedAt.UTC().Format(time.RFC3339),
			LastSeenAt:     session.LastSeenAt.UTC().Format(time.RFC3339),
		})
	}

	return &identityv1.ListSessionsResponse{Sessions: sessions}, nil
}

func (s *IdentityServer) RevokeSession(
	ctx context.Context,
	request *identityv1.RevokeSessionRequest,
) (*identityv1.RevokeSessionResponse, error) {
	userID, _, subject, err := s.signedInUser(ctx)
	if err != nil {
		return nil, err
	}

	target := request.GetAuth0SessionId()
	if target == "" {
		return nil, status.Error(codes.InvalidArgument, "a session id is required")
	}

	if err := s.registry.Revoke(ctx, userID, target); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identityv1.RevokeSessionResponse{Subject: subject}, nil
}
