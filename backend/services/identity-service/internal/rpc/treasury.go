package rpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type TreasuryDesignator interface {
	IssueNonce(ctx context.Context, userID uuid.UUID) (service.Nonce, error)
	Designate(
		ctx context.Context,
		organizationID, userID uuid.UUID,
		role domain.OrganizationRole,
		ownership service.Ownership,
	) (service.Designation, error)
}

func (s *IdentityServer) IssueTreasuryNonce(
	ctx context.Context,
	_ *identityv1.IssueTreasuryNonceRequest,
) (*identityv1.IssueTreasuryNonceResponse, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present {
		return nil, status.Error(codes.Unauthenticated, "a verified caller is required")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	nonce, err := s.treasury.IssueNonce(ctx, userID)
	if err != nil {
		s.logger.Error("issue treasury nonce failed", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "could not issue a nonce")
	}

	return &identityv1.IssueTreasuryNonceResponse{
		Nonce:     nonce.Value,
		Domain:    nonce.Domain,
		ChainId:   strconvItoa(nonce.ChainID),
		IssuedAt:  nonce.IssuedAt.UTC().Format(time.RFC3339),
		ExpiresAt: nonce.ExpiresAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *IdentityServer) DesignateTreasury(
	ctx context.Context,
	request *identityv1.DesignateTreasuryRequest,
) (*identityv1.DesignateTreasuryResponse, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return nil, status.Error(codes.Unauthenticated, "a verified organization is required")
	}

	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	key := grpcx.IdempotencyKeyFromIncoming(ctx)
	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	designation, err := s.treasury.Designate(ctx, organizationID, userID,
		domain.OrganizationRole(verified.Role),
		service.Ownership{
			Message:        request.GetMessage(),
			Signature:      request.GetSignature(),
			IdempotencyKey: key,
		})
	if err != nil {
		return nil, s.designationFailure(organizationID, err)
	}

	return &identityv1.DesignateTreasuryResponse{
		Address:      designation.Address,
		DesignatedAt: designation.DesignatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *IdentityServer) designationFailure(organizationID uuid.UUID, err error) error {
	switch {
	case errors.Is(err, service.ErrNotOrganizationOwner):
		return status.Error(codes.PermissionDenied, "OWNER_REQUIRED")
	case errors.Is(err, service.ErrProofRejected):
		return status.Error(codes.InvalidArgument, "PROOF_REJECTED")
	case errors.Is(err, service.ErrTreasuryDesignated):
		return status.Error(codes.AlreadyExists, "TREASURY_DESIGNATED")
	case errors.Is(err, service.ErrAddressTaken):
		return status.Error(codes.AlreadyExists, "ADDRESS_TAKEN")
	case errors.Is(err, service.ErrRequestInProgress):
		return status.Error(codes.Aborted, "REQUEST_IN_PROGRESS")
	case errors.Is(err, service.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, "IDEMPOTENCY_KEY_REUSED")
	}

	s.logger.Error("designate treasury failed",
		slog.String("organization_id", organizationID.String()),
		slog.Any("error", err),
	)
	return status.Error(codes.Internal, "could not designate the treasury address")
}

func strconvItoa(value int) string {
	return fmt.Sprintf("%d", value)
}
