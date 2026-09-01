package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

var ErrUnverifiedEmailClaim = errors.New("cannot link an unverified email to an existing user")

type Claims struct {
	Auth0Subject  string
	Email         string
	EmailVerified bool
	Name          string
}

type Session struct {
	User               domain.User
	Organization       *domain.Organization
	Role               domain.OrganizationRole
	TreasuryDesignated bool
}

func (s Session) NeedsOnboarding() bool { return s.Organization == nil }

type SessionService struct {
	users       repository.UserStore
	memberships repository.MembershipStore
	logger      *slog.Logger
}

func NewSessionService(
	users repository.UserStore,
	memberships repository.MembershipStore,
	logger *slog.Logger,
) *SessionService {
	return &SessionService{users: users, memberships: memberships, logger: logger}
}

func (s *SessionService) Resolve(ctx context.Context, claims Claims) (Session, error) {
	user, err := s.provision(ctx, claims)
	if err != nil {
		return Session{}, err
	}

	membership, err := s.memberships.FindActiveForUser(ctx, user.ID)
	if errors.Is(err, repository.ErrNoMembership) {
		return Session{User: user}, nil
	}
	if err != nil {
		return Session{}, err
	}

	snapshot, err := s.memberships.FindOrganization(ctx, membership.OrganizationID)
	if err != nil {
		return Session{}, err
	}

	return Session{
		User:               user,
		Organization:       &snapshot.Organization,
		Role:               membership.Role,
		TreasuryDesignated: snapshot.TreasuryDesignated,
	}, nil
}

func (s *SessionService) provision(ctx context.Context, claims Claims) (domain.User, error) {
	user, err := s.users.FindByAuth0Subject(ctx, claims.Auth0Subject)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, err
	}

	existing, err := s.users.FindByEmail(ctx, claims.Email)
	switch {
	case err == nil:
		return s.link(ctx, existing, claims)
	case errors.Is(err, repository.ErrUserNotFound):
		return s.create(ctx, claims)
	default:
		return domain.User{}, err
	}
}

func (s *SessionService) link(
	ctx context.Context,
	existing domain.User,
	claims Claims,
) (domain.User, error) {
	if !claims.EmailVerified {
		s.logger.Warn("refused to link unverified email to existing user",
			slog.String("subject", claims.Auth0Subject),
			slog.String("user_id", existing.ID.String()),
		)
		return domain.User{}, ErrUnverifiedEmailClaim
	}

	if err := s.users.AttachAuth0Subject(ctx, existing.ID, claims.Auth0Subject); err != nil {
		return domain.User{}, err
	}

	existing.Auth0Subject = &claims.Auth0Subject
	existing.EmailVerified = true
	return existing, nil
}

func (s *SessionService) create(ctx context.Context, claims Claims) (domain.User, error) {
	user := domain.User{
		Auth0Subject:  &claims.Auth0Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
	}

	err := s.users.Create(ctx, &user)
	if errors.Is(err, repository.ErrUserExists) {
		return s.recoverFromRace(ctx, claims)
	}
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *SessionService) recoverFromRace(ctx context.Context, claims Claims) (domain.User, error) {
	user, err := s.users.FindByAuth0Subject(ctx, claims.Auth0Subject)
	if err != nil {
		return domain.User{}, fmt.Errorf("concurrent provisioning conflict for %s: %w", claims.Auth0Subject, err)
	}
	return user, nil
}
