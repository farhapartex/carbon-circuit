package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

var (
	ErrNotPermitted            = errors.New("your role does not permit that")
	ErrOwnerRoleRequired       = errors.New("only an owner may grant or revoke the owner role")
	ErrLastOwner               = errors.New("an organization must always have at least one owner")
	ErrInvitationPending       = errors.New("an invitation is already pending for that email")
	ErrInvitationInvalid       = errors.New("invitation is not valid")
	ErrInvitationExpired       = errors.New("invitation has expired")
	ErrInvitationNotYours      = errors.New("invitation was issued to a different email address")
	ErrAlreadyInAnOrganization = errors.New("you already belong to an organization")
	ErrMemberUnknown           = errors.New("member not found")
)

type Team struct {
	Members     []repository.MemberRecord
	Invitations []domain.Invitation
}

type Invite struct {
	Email string
	Role  domain.OrganizationRole
}

type IssuedInvitation struct {
	Invitation domain.Invitation
	Token      string
}

type Accepted struct {
	OrganizationID   uuid.UUID
	OrganizationName string
	Role             domain.OrganizationRole
}

type Actor struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	Role           domain.OrganizationRole
}

func (a Actor) manages() bool {
	return a.Role == domain.RoleOwner || a.Role == domain.RoleAdmin
}

type TeamService struct {
	database      *gorm.DB
	team          repository.TeamStore
	users         repository.UserStore
	organizations repository.OrganizationReader
	existing      repository.MembershipStore
	invitationTTL time.Duration
	logger        *slog.Logger
}

func NewTeamService(
	handle *gorm.DB,
	team repository.TeamStore,
	users repository.UserStore,
	organizations repository.OrganizationReader,
	existing repository.MembershipStore,
	invitationTTL time.Duration,
	logger *slog.Logger,
) *TeamService {
	return &TeamService{
		database:      handle,
		team:          team,
		users:         users,
		organizations: organizations,
		existing:      existing,
		invitationTTL: invitationTTL,
		logger:        logger,
	}
}

func (s *TeamService) List(ctx context.Context, actor Actor) (Team, error) {
	var team Team

	err := s.within(ctx, actor, func(tx database.Tx) error {
		members, err := s.team.ListMembers(tx, actor.OrganizationID)
		if err != nil {
			return err
		}

		invitations, err := s.team.ListInvitations(tx, actor.OrganizationID)
		if err != nil {
			return err
		}

		team = Team{Members: members, Invitations: invitations}
		return nil
	})

	return team, err
}

func (s *TeamService) Invite(
	ctx context.Context,
	actor Actor,
	invite Invite,
) (IssuedInvitation, error) {
	if !actor.manages() {
		return IssuedInvitation{}, ErrNotPermitted
	}
	if invite.Role == domain.RoleOwner && actor.Role != domain.RoleOwner {
		return IssuedInvitation{}, ErrOwnerRoleRequired
	}

	token, hash, err := issueToken()
	if err != nil {
		return IssuedInvitation{}, err
	}

	invitation := domain.Invitation{
		OrganizationID:  actor.OrganizationID,
		Email:           strings.ToLower(strings.TrimSpace(invite.Email)),
		Role:            invite.Role,
		TokenHash:       hash,
		State:           domain.InvitationPending,
		InvitedByUserID: actor.UserID,
		ExpiresAt:       time.Now().Add(s.invitationTTL),
	}

	err = s.within(ctx, actor, func(tx database.Tx) error {
		if createErr := s.team.CreateInvitation(tx, &invitation); createErr != nil {
			if errors.Is(createErr, repository.ErrInvitationPending) {
				return ErrInvitationPending
			}
			return createErr
		}
		return nil
	})
	if err != nil {
		return IssuedInvitation{}, err
	}

	return IssuedInvitation{Invitation: invitation, Token: token}, nil
}

func (s *TeamService) RevokeInvitation(
	ctx context.Context,
	actor Actor,
	invitationID uuid.UUID,
) error {
	if !actor.manages() {
		return ErrNotPermitted
	}

	return s.within(ctx, actor, func(tx database.Tx) error {
		err := s.team.MarkInvitation(tx, invitationID, domain.InvitationRevoked, nil)
		if errors.Is(err, repository.ErrInvitationUnknown) {
			return ErrInvitationInvalid
		}
		return err
	})
}

func (s *TeamService) ChangeRole(
	ctx context.Context,
	actor Actor,
	subjectID uuid.UUID,
	role domain.OrganizationRole,
) (string, error) {
	if !actor.manages() {
		return "", ErrNotPermitted
	}

	var affected string

	err := s.within(ctx, actor, func(tx database.Tx) error {
		membership, err := s.team.FindMembership(tx, actor.OrganizationID, subjectID)
		if err != nil {
			if errors.Is(err, repository.ErrMemberUnknown) {
				return ErrMemberUnknown
			}
			return err
		}

		if role == domain.RoleOwner || membership.Role == domain.RoleOwner {
			if actor.Role != domain.RoleOwner {
				return ErrOwnerRoleRequired
			}
		}

		if membership.Role == domain.RoleOwner && role != domain.RoleOwner {
			if err := s.refuseLastOwner(tx, actor.OrganizationID); err != nil {
				return err
			}
		}

		if err := s.team.UpdateRole(tx, membership.ID, role); err != nil {
			return err
		}

		affected, err = s.team.FindSubject(tx, subjectID)
		return err
	})

	return affected, err
}

func (s *TeamService) RevokeMember(
	ctx context.Context,
	actor Actor,
	subjectID uuid.UUID,
) (string, error) {
	if !actor.manages() {
		return "", ErrNotPermitted
	}

	var affected string

	err := s.within(ctx, actor, func(tx database.Tx) error {
		membership, err := s.team.FindMembership(tx, actor.OrganizationID, subjectID)
		if err != nil {
			if errors.Is(err, repository.ErrMemberUnknown) {
				return ErrMemberUnknown
			}
			return err
		}

		if membership.Role == domain.RoleOwner {
			if actor.Role != domain.RoleOwner {
				return ErrOwnerRoleRequired
			}
			if err := s.refuseLastOwner(tx, actor.OrganizationID); err != nil {
				return err
			}
		}

		if err := s.team.RevokeMembership(tx, membership.ID, time.Now()); err != nil {
			return err
		}

		affected, err = s.team.FindSubject(tx, subjectID)
		return err
	})

	return affected, err
}

func (s *TeamService) refuseLastOwner(tx database.Tx, organizationID uuid.UUID) error {
	owners, err := s.team.CountActiveOwners(tx, organizationID)
	if err != nil {
		return err
	}
	if owners <= 1 {
		return ErrLastOwner
	}
	return nil
}

func (s *TeamService) within(
	ctx context.Context,
	actor Actor,
	work func(tx database.Tx) error,
) error {
	return database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         actor.UserID.String(),
			OrganizationID: actor.OrganizationID.String(),
		},
		work,
	)
}

func issueToken() (string, []byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("generate invitation token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	digest := sha256.Sum256([]byte(token))

	return token, digest[:], nil
}

func hashToken(token string) []byte {
	digest := sha256.Sum256([]byte(token))
	return digest[:]
}
