package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

func (s *TeamService) Accept(
	ctx context.Context,
	subject, token string,
) (Accepted, error) {
	user, err := s.users.FindByAuth0Subject(ctx, subject)
	if err != nil {
		return Accepted{}, err
	}

	invitation, err := s.locate(ctx, user, token)
	if err != nil {
		return Accepted{}, err
	}

	var accepted Accepted

	err = database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         user.ID.String(),
			OrganizationID: invitation.OrganizationID.String(),
		},
		func(tx database.Tx) error {
			organization, found, findErr := s.organizations.Find(tx, invitation.OrganizationID)
			if findErr != nil {
				return findErr
			}
			if !found {
				return ErrInvitationInvalid
			}

			joined := time.Now()
			membership := domain.OrganizationMembership{
				OrganizationID:  invitation.OrganizationID,
				UserID:          user.ID,
				Role:            invitation.Role,
				State:           domain.MembershipActive,
				InvitedByUserID: &invitation.InvitedByUserID,
				JoinedAt:        &joined,
			}

			if createErr := s.team.CreateMembership(tx, &membership); createErr != nil {
				if errors.Is(createErr, repository.ErrAlreadyMember) {
					return ErrAlreadyInAnOrganization
				}
				return createErr
			}

			if markErr := s.team.MarkInvitation(
				tx, invitation.ID, domain.InvitationAccepted, &user.ID,
			); markErr != nil {
				if errors.Is(markErr, repository.ErrInvitationUnknown) {
					return ErrInvitationInvalid
				}
				return markErr
			}

			accepted = Accepted{
				OrganizationID:   organization.ID,
				OrganizationName: organization.Name,
				Role:             invitation.Role,
			}
			return nil
		},
	)
	if err != nil {
		return Accepted{}, err
	}

	return accepted, nil
}

func (s *TeamService) locate(
	ctx context.Context,
	user domain.User,
	token string,
) (domain.Invitation, error) {
	if _, err := s.memberships(ctx, user.ID); err == nil {
		return domain.Invitation{}, ErrAlreadyInAnOrganization
	} else if !errors.Is(err, repository.ErrNoMembership) {
		return domain.Invitation{}, err
	}

	var invitation domain.Invitation

	err := database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{UserID: user.ID.String()},
		func(tx database.Tx) error {
			found, findErr := s.team.FindInvitationByHash(tx, hashToken(token))
			if findErr != nil {
				if errors.Is(findErr, repository.ErrInvitationUnknown) {
					return ErrInvitationInvalid
				}
				return findErr
			}
			invitation = found
			return nil
		},
	)
	if err != nil {
		return domain.Invitation{}, err
	}

	if invitation.State != domain.InvitationPending {
		return domain.Invitation{}, ErrInvitationInvalid
	}
	if time.Now().After(invitation.ExpiresAt) {
		return domain.Invitation{}, ErrInvitationExpired
	}
	if !strings.EqualFold(invitation.Email, user.Email) {
		return domain.Invitation{}, ErrInvitationNotYours
	}
	if !user.EmailVerified {
		return domain.Invitation{}, ErrInvitationNotYours
	}

	return invitation, nil
}

func (s *TeamService) memberships(
	ctx context.Context,
	userID uuid.UUID,
) (domain.OrganizationMembership, error) {
	return s.existing.FindActiveForUser(ctx, userID)
}
