package upstream

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
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
	connection, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(transport),
	)
	if err != nil {
		return nil, err
	}

	return &Identity{
		connection:  connection,
		client:      identityv1.NewIdentityServiceClient(connection),
		callTimeout: callTimeout,
	}, nil
}

func (i *Identity) Close() error {
	return i.connection.Close()
}

func (i *Identity) Ping(ctx context.Context) (*identityv1.PingResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))

	return i.client.Ping(callCtx, &identityv1.PingRequest{})
}

func (i *Identity) ResolveSession(
	ctx context.Context,
	caller auth.Caller,
) (*identityv1.ResolveSessionResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))

	return i.client.ResolveSession(callCtx, &identityv1.ResolveSessionRequest{
		Auth0Subject:  caller.Subject,
		Email:         caller.Email,
		EmailVerified: caller.EmailVerified,
		Name:          caller.Name,
	})
}

func (i *Identity) CreateOrganization(
	ctx context.Context,
	idempotencyKey string,
	request *identityv1.CreateOrganizationRequest,
) (*identityv1.CreateOrganizationResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))
	callCtx = grpcx.WithIdempotencyKey(callCtx, idempotencyKey)

	return i.client.CreateOrganization(callCtx, request)
}

func (i *Identity) GetOrganization(
	ctx context.Context,
) (*identityv1.GetOrganizationResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))

	return i.client.GetOrganization(callCtx, &identityv1.GetOrganizationRequest{})
}

func (i *Identity) IssueTreasuryNonce(
	ctx context.Context,
) (*identityv1.IssueTreasuryNonceResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))

	return i.client.IssueTreasuryNonce(callCtx, &identityv1.IssueTreasuryNonceRequest{})
}

func (i *Identity) DesignateTreasury(
	ctx context.Context,
	idempotencyKey string,
	message, signature string,
) (*identityv1.DesignateTreasuryResponse, error) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	defer cancel()

	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))
	callCtx = grpcx.WithIdempotencyKey(callCtx, idempotencyKey)

	return i.client.DesignateTreasury(callCtx, &identityv1.DesignateTreasuryRequest{
		Message:   message,
		Signature: signature,
	})
}

func (i *Identity) call(ctx context.Context, idempotencyKey string) (context.Context, context.CancelFunc) {
	callCtx, cancel := context.WithTimeout(ctx, i.callTimeout)
	callCtx = grpcx.WithCorrelationID(callCtx, logging.CorrelationIDFrom(ctx))
	callCtx = grpcx.WithIdempotencyKey(callCtx, idempotencyKey)
	return callCtx, cancel
}

func (i *Identity) ListMembers(ctx context.Context) (*identityv1.ListMembersResponse, error) {
	callCtx, cancel := i.call(ctx, "")
	defer cancel()
	return i.client.ListMembers(callCtx, &identityv1.ListMembersRequest{})
}

func (i *Identity) InviteMember(
	ctx context.Context,
	idempotencyKey, email string,
	role identityv1.OrganizationRole,
) (*identityv1.InviteMemberResponse, error) {
	callCtx, cancel := i.call(ctx, idempotencyKey)
	defer cancel()
	return i.client.InviteMember(callCtx, &identityv1.InviteMemberRequest{
		Email: email,
		Role:  role,
	})
}

func (i *Identity) RevokeInvitation(
	ctx context.Context,
	idempotencyKey, invitationID string,
) error {
	callCtx, cancel := i.call(ctx, idempotencyKey)
	defer cancel()
	_, err := i.client.RevokeInvitation(callCtx, &identityv1.RevokeInvitationRequest{
		InvitationId: invitationID,
	})
	return err
}

func (i *Identity) ChangeMemberRole(
	ctx context.Context,
	idempotencyKey, userID string,
	role identityv1.OrganizationRole,
) (*identityv1.ChangeMemberRoleResponse, error) {
	callCtx, cancel := i.call(ctx, idempotencyKey)
	defer cancel()
	return i.client.ChangeMemberRole(callCtx, &identityv1.ChangeMemberRoleRequest{
		UserId: userID,
		Role:   role,
	})
}

func (i *Identity) RevokeMember(
	ctx context.Context,
	idempotencyKey, userID string,
) (*identityv1.RevokeMemberResponse, error) {
	callCtx, cancel := i.call(ctx, idempotencyKey)
	defer cancel()
	return i.client.RevokeMember(callCtx, &identityv1.RevokeMemberRequest{UserId: userID})
}

func (i *Identity) AcceptInvitation(
	ctx context.Context,
	idempotencyKey, token string,
) (*identityv1.AcceptInvitationResponse, error) {
	callCtx, cancel := i.call(ctx, idempotencyKey)
	defer cancel()
	return i.client.AcceptInvitation(callCtx, &identityv1.AcceptInvitationRequest{Token: token})
}

func (i *Identity) CreateFacility(
	ctx context.Context,
	idempotencyKey string,
	request *identityv1.CreateFacilityRequest,
) (*identityv1.CreateFacilityResponse, error) {
	callCtx, cancel := i.call(ctx, idempotencyKey)
	defer cancel()
	return i.client.CreateFacility(callCtx, request)
}

func (i *Identity) ListFacilities(
	ctx context.Context,
) (*identityv1.ListFacilitiesResponse, error) {
	callCtx, cancel := i.call(ctx, "")
	defer cancel()
	return i.client.ListFacilities(callCtx, &identityv1.ListFacilitiesRequest{})
}

func (i *Identity) GetFacility(
	ctx context.Context,
	facilityID string,
) (*identityv1.GetFacilityResponse, error) {
	callCtx, cancel := i.call(ctx, "")
	defer cancel()
	return i.client.GetFacility(callCtx, &identityv1.GetFacilityRequest{
		FacilityId: facilityID,
	})
}
