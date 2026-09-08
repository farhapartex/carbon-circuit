package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/apikey"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

func apiKeyService(t *testing.T, handle *gorm.DB) *service.APIKeyService {
	t.Helper()

	hasher, err := apikey.NewHasher("integration-test-pepper")
	if err != nil {
		t.Fatalf("build hasher: %v", err)
	}

	return service.NewAPIKeyService(
		handle,
		repository.NewAPIKeyRepository(),
		repository.NewOrganizationRepository(),
		hasher,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func seedKeyOrganization(t *testing.T, handle *gorm.DB) service.Actor {
	t.Helper()

	userID, _ := seedUser(t, handle)
	organizationID, _ := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationVerified,
	)

	t.Cleanup(func() {
		scoped := database.TenantContext{
			UserID:         userID.String(),
			OrganizationID: organizationID.String(),
		}
		err := database.WithinTenant(context.Background(), handle, scoped,
			func(tx database.Tx) error {
				return tx.Session().Exec(
					`DELETE FROM identity.api_keys WHERE organization_id = ?`,
					organizationID,
				).Error
			})
		if err != nil {
			t.Errorf("clean api keys for %s: %v", organizationID, err)
		}
	})

	return service.Actor{
		OrganizationID: organizationID,
		UserID:         userID,
		Role:           domain.RoleOwner,
	}
}

func TestCreatedKeyIsPresentedOnceAndStoredHashed(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "ERP checkpoint ingest")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	if !strings.HasPrefix(issued.Presented, apikey.Scheme+"_") {
		t.Fatalf("expected a scheme prefixed key, got %q", issued.Presented)
	}
	if !strings.Contains(issued.Presented, issued.Key.Prefix) {
		t.Fatal("the presented key must contain its lookup prefix")
	}

	listed, err := keys.List(context.Background(), actor)
	if err != nil {
		t.Fatalf("list keys: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one key, got %d", len(listed))
	}

	stored := listed[0]
	if len(stored.SecretHMAC) != 32 {
		t.Fatalf("expected a 32 byte stored hash, got %d", len(stored.SecretHMAC))
	}
	if strings.Contains(string(stored.SecretHMAC), issued.Presented) {
		t.Fatal("the stored hash must not contain the presented key")
	}
}

func TestListedKeysNeverCarryARecoverableSecret(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "Warehouse bridge")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	listed, err := keys.List(context.Background(), actor)
	if err != nil {
		t.Fatalf("list keys: %v", err)
	}

	parsed, err := apikey.Parse(issued.Presented)
	if err != nil {
		t.Fatalf("parse issued key: %v", err)
	}

	for _, key := range listed {
		if strings.Contains(string(key.SecretHMAC), parsed.Secret) {
			t.Fatal("a listed key exposed its secret")
		}
	}
}

func TestRevokingAKeyLeavesItListedAsRevoked(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "Legacy bridge")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	if err := keys.Revoke(context.Background(), actor, issued.Key.ID); err != nil {
		t.Fatalf("revoke key: %v", err)
	}

	listed, err := keys.List(context.Background(), actor)
	if err != nil {
		t.Fatalf("list keys: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("a revoked key must stay on the record, got %d rows", len(listed))
	}
	if listed[0].Active() {
		t.Fatal("expected the key to read as revoked")
	}
	if listed[0].RevokedByUserID == nil {
		t.Fatal("expected the revoking user to be recorded")
	}
}

func TestRevokingTwiceIsRefused(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "Bridge")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	if err := keys.Revoke(context.Background(), actor, issued.Key.ID); err != nil {
		t.Fatalf("first revoke: %v", err)
	}

	if err := keys.Revoke(
		context.Background(), actor, issued.Key.ID,
	); !errors.Is(err, service.ErrAPIKeyNotFound) {
		t.Fatalf("expected ErrAPIKeyNotFound, got %v", err)
	}
}

func TestMemberCannotCreateOrRevokeKeys(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "Owner key")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	member := actor
	member.Role = domain.RoleMember

	if _, err := keys.Create(
		context.Background(), member, "Member key",
	); !errors.Is(err, service.ErrNotPermitted) {
		t.Fatalf("expected ErrNotPermitted on create, got %v", err)
	}

	if err := keys.Revoke(
		context.Background(), member, issued.Key.ID,
	); !errors.Is(err, service.ErrNotPermitted) {
		t.Fatalf("expected ErrNotPermitted on revoke, got %v", err)
	}
}

func TestAnotherOrganizationCannotSeeOrRevokeYourKeys(t *testing.T) {
	handle := store(t)
	owner := seedKeyOrganization(t, handle)
	stranger := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), owner, "Owner key")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	listed, err := keys.List(context.Background(), stranger)
	if err != nil {
		t.Fatalf("list as stranger: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("a stranger saw %d keys that are not theirs", len(listed))
	}

	if err := keys.Revoke(
		context.Background(), stranger, issued.Key.ID,
	); !errors.Is(err, service.ErrAPIKeyNotFound) {
		t.Fatalf("a stranger must not revoke another organization's key, got %v", err)
	}
}

func TestKeyNeedsAName(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	for _, name := range []string{"", "   ", strings.Repeat("x", 200)} {
		if _, err := keys.Create(
			context.Background(), actor, name,
		); !errors.Is(err, service.ErrAPIKeyNameEmpty) {
			t.Fatalf("expected ErrAPIKeyNameEmpty for %q, got %v", name, err)
		}
	}
}

func TestValidateAcceptsTheIssuedKeyAndReturnsItsOrganization(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "ERP ingest")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	validated, err := keys.Validate(context.Background(), issued.Presented)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}

	if validated.OrganizationID != actor.OrganizationID {
		t.Fatalf("expected the owning organization, got %s", validated.OrganizationID)
	}
	if validated.ActingUserID != actor.UserID {
		t.Fatalf("expected the key to act as its creator, got %s", validated.ActingUserID)
	}
	if validated.Prefix != issued.Key.Prefix {
		t.Fatalf("expected the prefix to round trip, got %q", validated.Prefix)
	}
}

func TestValidateRefusesTamperedAndUnknownKeys(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "ERP ingest")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	cases := map[string]string{
		"empty":          "",
		"not a key":      "hello",
		"unknown prefix": "cc_live_zzzzzzzz_" + strings.Repeat("a", 43),
		"secret mutated": issued.Presented + "x",
		"prefix mutated": strings.Replace(issued.Presented, issued.Key.Prefix, "aaaaaaaa", 1),
	}

	for name, candidate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := keys.Validate(context.Background(), candidate); err == nil {
				t.Fatalf("expected %s to be refused", name)
			}
		})
	}
}

func TestValidateRefusesARevokedKey(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "Legacy bridge")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	if err := keys.Revoke(context.Background(), actor, issued.Key.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	if _, err := keys.Validate(
		context.Background(), issued.Presented,
	); !errors.Is(err, service.ErrAPIKeyRevoked) {
		t.Fatalf("expected ErrAPIKeyRevoked, got %v", err)
	}
}

func TestValidateRecordsThatTheKeyWasUsed(t *testing.T) {
	handle := store(t)
	actor := seedKeyOrganization(t, handle)
	keys := apiKeyService(t, handle)

	issued, err := keys.Create(context.Background(), actor, "ERP ingest")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	before, err := keys.List(context.Background(), actor)
	if err != nil {
		t.Fatalf("list before: %v", err)
	}
	if before[0].LastUsedAt != nil {
		t.Fatal("a new key should not report use")
	}

	if _, err := keys.Validate(context.Background(), issued.Presented); err != nil {
		t.Fatalf("validate: %v", err)
	}

	after, err := keys.List(context.Background(), actor)
	if err != nil {
		t.Fatalf("list after: %v", err)
	}
	if after[0].LastUsedAt == nil {
		t.Fatal("expected the key to record when it was used")
	}
}
