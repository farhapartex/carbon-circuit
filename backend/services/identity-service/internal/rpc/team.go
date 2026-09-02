package rpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

type TeamManager interface {
	List(ctx context.Context, actor service.Actor) (service.Team, error)
	Invite(ctx context.Context, actor service.Actor, invite service.Invite) (service.IssuedInvitation, error)
	RevokeInvitation(ctx context.Context, actor service.Actor, invitationID uuid.UUID) error
	ChangeRole(ctx context.Context, actor service.Actor, subjectID uuid.UUID, role domain.OrganizationRole) (string, error)
	RevokeMember(ctx context.Context, actor service.Actor, subjectID uuid.UUID) (string, error)
	Accept(ctx context.Context, subject, token string) (service.Accepted, error)
}

var invitationStates = map[domain.InvitationState]identityv1.InvitationState{
	domain.InvitationPending:  identityv1.InvitationState_INVITATION_STATE_PENDING,
	domain.InvitationAccepted: identityv1.InvitationState_INVITATION_STATE_ACCEPTED,
	domain.InvitationRevoked:  identityv1.InvitationState_INVITATION_STATE_REVOKED,
	domain.InvitationExpired:  identityv1.InvitationState_INVITATION_STATE_EXPIRED,
}

var rolesByName = map[identityv1.OrganizationRole]domain.OrganizationRole{
	identityv1.OrganizationRole_ORGANIZATION_ROLE_OWNER:  domain.RoleOwner,
	identityv1.OrganizationRole_ORGANIZATION_ROLE_ADMIN:  domain.RoleAdmin,
	identityv1.OrganizationRole_ORGANIZATION_ROLE_MEMBER: domain.RoleMember,
}

func actorFrom(verified servicetoken.Caller) (service.Actor, error) {
	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	return service.Actor{
		OrganizationID: organizationID,
		UserID:         userID,
		Role:           domain.OrganizationRole(verified.Role),
	}, nil
}

func (s *IdentityServer) actor(ctx context.Context) (service.Actor, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return service.Actor{}, status.Error(codes.Unauthenticated, "a verified organization is required")
	}
	return actorFrom(verified)
}

func (s *IdentityServer) ListMembers(
	ctx context.Context,
	_ *identityv1.ListMembersRequest,
) (*identityv1.ListMembersResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	team, err := s.team.List(ctx, actor)
	if err != nil {
		return nil, s.teamFailure(actor, err)
	}

	response := &identityv1.ListMembersResponse{
		Members:     make([]*identityv1.Member, 0, len(team.Members)),
		Invitations: make([]*identityv1.Invitation, 0, len(team.Invitations)),
	}

	for _, member := range team.Members {
		response.Members = append(response.Members, &identityv1.Member{
			UserId:       member.UserID.String(),
			Email:        member.Email,
			Name:         member.Name,
			Role:         organizationRoles[member.Role],
			MfaEnrolled:  member.MFAEnrolled,
			JoinedAt:     timestamp(member.JoinedAt),
			LastActiveAt: timestamp(member.LastActiveAt),
		})
	}

	for _, invitation := range team.Invitations {
		response.Invitations = append(response.Invitations, invitationToProto(invitation))
	}

	return response, nil
}

func (s *IdentityServer) InviteMember(
	ctx context.Context,
	request *identityv1.InviteMemberRequest,
) (*identityv1.InviteMemberResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	role, known := rolesByName[request.GetRole()]
	if !known {
		return nil, status.Error(codes.InvalidArgument, "role must be a known organization role")
	}

	email := strings.ToLower(strings.TrimSpace(request.GetEmail()))
	if email == "" || !strings.Contains(email, "@") {
		return nil, status.Error(codes.InvalidArgument, "a valid email is required")
	}

	issued, err := s.team.Invite(ctx, actor, service.Invite{Email: email, Role: role})
	if err != nil {
		return nil, s.teamFailure(actor, err)
	}

	return &identityv1.InviteMemberResponse{
		Invitation: invitationToProto(issued.Invitation),
		Token:      issued.Token,
	}, nil
}

func (s *IdentityServer) RevokeInvitation(
	ctx context.Context,
	request *identityv1.RevokeInvitationRequest,
) (*identityv1.RevokeInvitationResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	invitationID, err := uuid.Parse(request.GetInvitationId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invitation_id must be a uuid")
	}

	if err := s.team.RevokeInvitation(ctx, actor, invitationID); err != nil {
		return nil, s.teamFailure(actor, err)
	}

	return &identityv1.RevokeInvitationResponse{}, nil
}

func (s *IdentityServer) ChangeMemberRole(
	ctx context.Context,
	request *identityv1.ChangeMemberRoleRequest,
) (*identityv1.ChangeMemberRoleResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	subjectID, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a uuid")
	}

	role, known := rolesByName[request.GetRole()]
	if !known {
		return nil, status.Error(codes.InvalidArgument, "role must be a known organization role")
	}

	affected, err := s.team.ChangeRole(ctx, actor, subjectID, role)
	if err != nil {
		return nil, s.teamFailure(actor, err)
	}

	return &identityv1.ChangeMemberRoleResponse{
		Member:          &identityv1.Member{UserId: subjectID.String(), Role: organizationRoles[role]},
		AffectedSubject: affected,
	}, nil
}

func (s *IdentityServer) RevokeMember(
	ctx context.Context,
	request *identityv1.RevokeMemberRequest,
) (*identityv1.RevokeMemberResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	subjectID, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id must be a uuid")
	}

	affected, err := s.team.RevokeMember(ctx, actor, subjectID)
	if err != nil {
		return nil, s.teamFailure(actor, err)
	}

	return &identityv1.RevokeMemberResponse{AffectedSubject: affected}, nil
}

func (s *IdentityServer) AcceptInvitation(
	ctx context.Context,
	request *identityv1.AcceptInvitationRequest,
) (*identityv1.AcceptInvitationResponse, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || verified.Subject == "" {
		return nil, status.Error(codes.Unauthenticated, "a verified caller is required")
	}

	token := strings.TrimSpace(request.GetToken())
	if token == "" {
		return nil, status.Error(codes.InvalidArgument, "a token is required")
	}

	accepted, err := s.team.Accept(ctx, verified.Subject, token)
	if err != nil {
		return nil, s.acceptFailure(verified.Subject, err)
	}

	return &identityv1.AcceptInvitationResponse{
		OrganizationId:   accepted.OrganizationID.String(),
		OrganizationName: accepted.OrganizationName,
		Role:             organizationRoles[accepted.Role],
	}, nil
}

func invitationToProto(invitation domain.Invitation) *identityv1.Invitation {
	return &identityv1.Invitation{
		Id:        invitation.ID.String(),
		Email:     invitation.Email,
		Role:      organizationRoles[invitation.Role],
		State:     invitationStates[invitation.State],
		InvitedAt: invitation.CreatedAt.UTC().Format(time.RFC3339),
		ExpiresAt: invitation.ExpiresAt.UTC().Format(time.RFC3339),
	}
}

func timestamp(at *time.Time) string {
	if at == nil {
		return ""
	}
	return at.UTC().Format(time.RFC3339)
}

func (s *IdentityServer) teamFailure(actor service.Actor, err error) error {
	switch {
	case errors.Is(err, service.ErrNotPermitted):
		return status.Error(codes.PermissionDenied, "ROLE_NOT_PERMITTED")
	case errors.Is(err, service.ErrOwnerRoleRequired):
		return status.Error(codes.PermissionDenied, "OWNER_REQUIRED")
	case errors.Is(err, service.ErrLastOwner):
		return status.Error(codes.FailedPrecondition, "LAST_OWNER")
	case errors.Is(err, service.ErrInvitationPending):
		return status.Error(codes.AlreadyExists, "INVITATION_PENDING")
	case errors.Is(err, service.ErrInvitationInvalid):
		return status.Error(codes.NotFound, "INVITATION_INVALID")
	case errors.Is(err, service.ErrMemberUnknown):
		return status.Error(codes.NotFound, "MEMBER_NOT_FOUND")
	}

	s.logger.Error("team operation failed",
		slog.String("organization_id", actor.OrganizationID.String()),
		slog.Any("error", err),
	)
	return status.Error(codes.Internal, "could not complete the request")
}

func (s *IdentityServer) acceptFailure(subject string, err error) error {
	switch {
	case errors.Is(err, service.ErrInvitationInvalid):
		return status.Error(codes.NotFound, "INVITATION_INVALID")
	case errors.Is(err, service.ErrInvitationExpired):
		return status.Error(codes.FailedPrecondition, "INVITATION_EXPIRED")
	case errors.Is(err, service.ErrInvitationNotYours):
		return status.Error(codes.PermissionDenied, "INVITATION_NOT_YOURS")
	case errors.Is(err, service.ErrAlreadyInAnOrganization):
		return status.Error(codes.AlreadyExists, "ALREADY_IN_ORGANIZATION")
	case errors.Is(err, repository.ErrUserNotFound):
		return status.Error(codes.NotFound, "USER_NOT_FOUND")
	}

	s.logger.Error("accept invitation failed",
		slog.String("subject", subject),
		slog.Any("error", err),
	)
	return status.Error(codes.Internal, "could not accept the invitation")
}
