package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

type fakeUsers struct {
	bySubject map[string]domain.User
	byEmail   map[string]domain.User
	created   []domain.User
	attached  map[uuid.UUID]string
	createErr error
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		bySubject: map[string]domain.User{},
		byEmail:   map[string]domain.User{},
		attached:  map[uuid.UUID]string{},
	}
}

func (f *fakeUsers) FindByAuth0Subject(_ context.Context, subject string) (domain.User, error) {
	if user, found := f.bySubject[subject]; found {
		return user, nil
	}
	return domain.User{}, repository.ErrUserNotFound
}

func (f *fakeUsers) FindByEmail(_ context.Context, email string) (domain.User, error) {
	if user, found := f.byEmail[email]; found {
		return user, nil
	}
	return domain.User{}, repository.ErrUserNotFound
}

func (f *fakeUsers) Create(_ context.Context, user *domain.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	user.ID = uuid.New()
	f.created = append(f.created, *user)
	return nil
}

func (f *fakeUsers) AttachAuth0Subject(_ context.Context, userID uuid.UUID, subject string) error {
	f.attached[userID] = subject
	return nil
}

type fakeMemberships struct {
	membership   *domain.OrganizationMembership
	organization domain.Organization
}

func (f *fakeMemberships) FindActiveForUser(
	_ context.Context,
	_ uuid.UUID,
) (domain.OrganizationMembership, error) {
	if f.membership == nil {
		return domain.OrganizationMembership{}, repository.ErrNoMembership
	}
	return *f.membership, nil
}

func (f *fakeMemberships) FindOrganization(
	_ context.Context,
	_ uuid.UUID,
) (domain.Organization, error) {
	return f.organization, nil
}

func newService(users repository.UserStore, memberships repository.MembershipStore) *SessionService {
	return NewSessionService(users, memberships, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func verifiedClaims() Claims {
	return Claims{
		Auth0Subject:  "google-oauth2|1234",
		Email:         "nazmul@example.test",
		EmailVerified: true,
		Name:          "Nazmul",
	}
}

func TestKnownSubjectIsReturnedWithoutWrites(t *testing.T) {
	users := newFakeUsers()
	existing := domain.User{Email: "nazmul@example.test", Name: "Nazmul"}
	existing.ID = uuid.New()
	users.bySubject["google-oauth2|1234"] = existing

	session, err := newService(users, &fakeMemberships{}).Resolve(context.Background(), verifiedClaims())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if session.User.ID != existing.ID {
		t.Fatal("expected the existing user to be returned")
	}
	if len(users.created) != 0 || len(users.attached) != 0 {
		t.Fatal("expected no writes for a known subject")
	}
}

func TestUnknownSubjectAndEmailCreatesUser(t *testing.T) {
	users := newFakeUsers()

	session, err := newService(users, &fakeMemberships{}).Resolve(context.Background(), verifiedClaims())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if len(users.created) != 1 {
		t.Fatalf("expected one user to be created, got %d", len(users.created))
	}
	if !session.NeedsOnboarding() {
		t.Fatal("a brand new user with no membership must need onboarding")
	}
}

func TestVerifiedEmailLinksToExistingUser(t *testing.T) {
	users := newFakeUsers()
	invited := domain.User{Email: "nazmul@example.test", Name: "Nazmul"}
	invited.ID = uuid.New()
	users.byEmail["nazmul@example.test"] = invited

	session, err := newService(users, &fakeMemberships{}).Resolve(context.Background(), verifiedClaims())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if users.attached[invited.ID] != "google-oauth2|1234" {
		t.Fatal("expected the auth0 subject to be attached to the invited user")
	}
	if len(users.created) != 0 {
		t.Fatal("expected no duplicate user to be created")
	}
	if session.User.ID != invited.ID {
		t.Fatal("expected the existing user identity to be preserved")
	}
}

func TestUnverifiedEmailIsRefusedInsteadOfLinking(t *testing.T) {
	users := newFakeUsers()
	victim := domain.User{Email: "nazmul@example.test", Name: "Nazmul"}
	victim.ID = uuid.New()
	users.byEmail["nazmul@example.test"] = victim

	claims := verifiedClaims()
	claims.EmailVerified = false
	claims.Auth0Subject = "attacker|9999"

	_, err := newService(users, &fakeMemberships{}).Resolve(context.Background(), claims)
	if !errors.Is(err, ErrUnverifiedEmailClaim) {
		t.Fatalf("expected ErrUnverifiedEmailClaim, got %v", err)
	}

	if _, attached := users.attached[victim.ID]; attached {
		t.Fatal("an unverified email must never be linked to an existing user")
	}
	if len(users.created) != 0 {
		t.Fatal("expected no user to be created on a refused link")
	}
}

func TestConcurrentProvisioningRecoversFromRace(t *testing.T) {
	users := newFakeUsers()
	users.createErr = repository.ErrUserExists

	winner := domain.User{Email: "nazmul@example.test", Name: "Nazmul"}
	winner.ID = uuid.New()

	service := newService(users, &fakeMemberships{})

	if _, err := service.Resolve(context.Background(), verifiedClaims()); err == nil {
		t.Fatal("expected an error while the winning row is not yet visible")
	}

	users.bySubject["google-oauth2|1234"] = winner

	session, err := service.Resolve(context.Background(), verifiedClaims())
	if err != nil {
		t.Fatalf("resolve after race: %v", err)
	}
	if session.User.ID != winner.ID {
		t.Fatal("expected the winning row to be adopted")
	}
}

func TestActiveMembershipResolvesOrganizationAndRole(t *testing.T) {
	users := newFakeUsers()
	existing := domain.User{Email: "nazmul@example.test"}
	existing.ID = uuid.New()
	users.bySubject["google-oauth2|1234"] = existing

	organization := domain.Organization{
		Name:               "Alice Corp",
		Type:               domain.OrganizationManufacturer,
		State:              domain.OrganizationActive,
		VerificationStatus: domain.VerificationUnverified,
	}
	organization.ID = uuid.New()

	membership := domain.OrganizationMembership{
		OrganizationID: organization.ID,
		UserID:         existing.ID,
		Role:           domain.RoleOwner,
		State:          domain.MembershipActive,
	}

	memberships := &fakeMemberships{membership: &membership, organization: organization}

	session, err := newService(users, memberships).Resolve(context.Background(), verifiedClaims())
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if session.NeedsOnboarding() {
		t.Fatal("a user with an active membership must not need onboarding")
	}
	if session.Organization == nil || session.Organization.Name != "Alice Corp" {
		t.Fatal("expected the organization to be resolved")
	}
	if session.Role != domain.RoleOwner {
		t.Fatalf("expected owner role, got %q", session.Role)
	}
}
