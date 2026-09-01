package service_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

func store(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run organization integration tests")
	}

	opened, err := database.Open(context.Background(), database.Options{
		DSN:             dsn,
		Schema:          "identity",
		MaxOpenConns:    8,
		MaxIdleConns:    4,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Minute,
		AcquireTimeout:  5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	return opened
}

func organizations(handle *gorm.DB) *service.OrganizationService {
	users := repository.NewUserRepository(handle)
	memberships := repository.NewMembershipRepository(handle)
	writer := repository.NewOrganizationRepository()

	return service.NewOrganizationService(
		handle, users, memberships, writer, writer,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func seedUser(t *testing.T, handle *gorm.DB) (uuid.UUID, string) {
	t.Helper()

	marker := uuid.New()
	subject := "probe|" + marker.String()

	err := handle.Exec(
		`INSERT INTO identity.users (auth0_subject, email, name, email_verified)
		 VALUES (?, ?, 'Probe User', true)`,
		subject, marker.String()+"@probe.test",
	).Error
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	var raw string
	if err := handle.Raw(
		`SELECT id::text FROM identity.users WHERE auth0_subject = ?`, subject,
	).Scan(&raw).Error; err != nil {
		t.Fatalf("read seeded user: %v", err)
	}

	userID := uuid.MustParse(raw)

	t.Cleanup(func() {
		scoped := database.TenantContext{UserID: userID.String()}

		err := database.WithinTenant(context.Background(), handle, scoped, func(tx database.Tx) error {
			if err := tx.Session().Exec(
				`DELETE FROM identity.idempotency_records WHERE user_id = ?`, userID).Error; err != nil {
				return err
			}
			return tx.Session().Exec(
				`DELETE FROM identity.organization_memberships WHERE user_id = ?`, userID).Error
		})
		if err != nil {
			t.Errorf("clear tenant rows for %s: %v", userID, err)
			return
		}

		if err := handle.Exec(`DELETE FROM identity.users WHERE id = ?`, userID).Error; err != nil {
			t.Errorf("delete probe user %s: %v", userID, err)
		}
	})

	return userID, subject
}

func registryRow(t *testing.T, handle *gorm.DB, condition string) (string, string, string) {
	t.Helper()

	var row struct {
		CountryCode        string
		RegistrationNumber string
		LegalName          string
	}

	query := fmt.Sprintf(
		`SELECT country_code, registration_number, legal_name
		 FROM identity.business_registry_records WHERE %s LIMIT 1`, condition,
	)
	if err := handle.Raw(query).Scan(&row).Error; err != nil {
		t.Fatalf("select registry row: %v", err)
	}
	if row.RegistrationNumber == "" {
		t.Fatalf("no registry row satisfies %s", condition)
	}

	return row.CountryCode, row.RegistrationNumber, row.LegalName
}

func retire(t *testing.T, handle *gorm.DB, organizationID uuid.UUID) {
	t.Helper()

	t.Cleanup(func() {
		handle.Exec(`DELETE FROM identity.outbox_events WHERE aggregate_id = ?`, organizationID)

		err := database.WithinTenant(
			context.Background(),
			handle,
			database.TenantContext{OrganizationID: organizationID.String()},
			func(tx database.Tx) error {
				return tx.Session().Exec(
					`UPDATE identity.organizations SET deleted_at = now() WHERE id = ?`,
					organizationID,
				).Error
			},
		)
		if err != nil {
			t.Errorf("retire organization %s: %v", organizationID, err)
		}
	})
}

func create(
	t *testing.T,
	handle *gorm.DB,
	creator *service.OrganizationService,
	subject string,
	request service.Registration,
) (service.Registered, error) {
	t.Helper()

	registered, err := creator.Create(context.Background(), subject, request)
	if err == nil {
		retire(t, handle, registered.Organization.ID)
	}
	return registered, err
}

func registration(name, country, number, key string) service.Registration {
	return service.Registration{
		LegalName:          name,
		Type:               domain.OrganizationManufacturer,
		CountryCode:        country,
		RegistrationNumber: number,
		ProductCategories:  []string{"electronics"},
		IdempotencyKey:     key,
		RequestBody:        []byte(name + "|" + country + "|" + number),
	}
}

func TestExactRegistryMatchVerifies(t *testing.T) {
	handle := store(t)
	_, subject := seedUser(t, handle)
	country, number, legalName := registryRow(t, handle, "entity_status = 'active' AND NOT sanctioned")

	registered, err := create(t, handle, organizations(handle), subject, registration(legalName, country, number, "key-verified-1"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if registered.Outcome.Status != domain.VerificationVerified {
		t.Fatalf("expected verified, got %q", registered.Outcome.Status)
	}
	if registered.Organization.State != domain.OrganizationActive {
		t.Fatalf("expected an active organization, got %q", registered.Organization.State)
	}
	if registered.Role != domain.RoleOwner {
		t.Fatalf("expected the creator to be owner, got %q", registered.Role)
	}

	var events int64
	handle.Table("identity.outbox_events").
		Where("aggregate_id = ? AND event_type = 'organization.verified'", registered.Organization.ID).
		Count(&events)
	if events != 1 {
		t.Fatalf("expected one organization.verified event, got %d", events)
	}
}

func TestUnknownRegistrationNumberIsUnverified(t *testing.T) {
	handle := store(t)
	_, subject := seedUser(t, handle)

	registered, err := create(t, handle, organizations(handle), subject, registration("Nowhere Industries Ltd.", "TW", "TW-00000000", "key-unverified-1"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if registered.Outcome.Status != domain.VerificationUnverified {
		t.Fatalf("expected unverified, got %q", registered.Outcome.Status)
	}
	if registered.Organization.State != domain.OrganizationActive {
		t.Fatal("an unverified organization must remain active, not suspended")
	}
	if registered.Outcome.MatchFound {
		t.Fatal("expected no registry match")
	}

	var events int64
	handle.Table("identity.outbox_events").
		Where("aggregate_id = ?", registered.Organization.ID).Count(&events)
	if events != 0 {
		t.Fatalf("expected no event for an unverified organization, got %d", events)
	}
}

func TestDissolvedEntityIsRejectedAndSuspended(t *testing.T) {
	handle := store(t)
	_, subject := seedUser(t, handle)
	country, number, legalName := registryRow(t, handle, "entity_status = 'dissolved'")

	registered, err := create(t, handle, organizations(handle), subject, registration(legalName, country, number, "key-dissolved-1"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if registered.Outcome.Status != domain.VerificationRejected {
		t.Fatalf("expected rejected, got %q", registered.Outcome.Status)
	}
	if registered.Organization.State != domain.OrganizationSuspended {
		t.Fatalf("expected suspended, got %q", registered.Organization.State)
	}
	if registered.Outcome.Rejection == nil || *registered.Outcome.Rejection != domain.RejectionEntityDissolved {
		t.Fatal("expected an entity_dissolved rejection")
	}
}

func TestSanctionedEntityIsRejectedAndSuspended(t *testing.T) {
	handle := store(t)
	_, subject := seedUser(t, handle)
	country, number, legalName := registryRow(t, handle, "sanctioned AND entity_status = 'active'")

	registered, err := create(t, handle, organizations(handle), subject, registration(legalName, country, number, "key-sanctioned-1"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if registered.Organization.State != domain.OrganizationSuspended {
		t.Fatalf("expected suspended, got %q", registered.Organization.State)
	}
	if registered.Outcome.Rejection == nil || *registered.Outcome.Rejection != domain.RejectionSanctionsFlag {
		t.Fatal("expected a sanctions_flag rejection")
	}
}

func TestNameMismatchIsRejected(t *testing.T) {
	handle := store(t)
	_, subject := seedUser(t, handle)
	country, number, _ := registryRow(t, handle, "entity_status = 'active' AND NOT sanctioned")

	registered, err := create(t, handle, organizations(handle), subject, registration("Totally Unrelated Trading Company", country, number, "key-mismatch-1"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if registered.Outcome.Rejection == nil || *registered.Outcome.Rejection != domain.RejectionNameMismatch {
		t.Fatalf("expected a name_mismatch rejection, got %v", registered.Outcome.Rejection)
	}
	if registered.Organization.State != domain.OrganizationSuspended {
		t.Fatal("a name mismatch must suspend the organization")
	}
	if registered.Outcome.NameSimilarity == nil {
		t.Fatal("expected the similarity score to be recorded")
	}
}

func TestRetryWithSameKeyReplaysWithoutCreatingAnother(t *testing.T) {
	handle := store(t)
	userID, subject := seedUser(t, handle)
	country, number, legalName := registryRow(t, handle, "entity_status = 'active' AND NOT sanctioned")

	creator := organizations(handle)
	request := registration(legalName, country, number, "key-replay-1")

	first, err := create(t, handle, creator, subject, request)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	second, err := creator.Create(context.Background(), subject, request)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}

	if !second.Replayed {
		t.Fatal("expected the retry to replay the stored response")
	}
	if second.Organization.ID != first.Organization.ID {
		t.Fatal("expected the replay to return the original organization")
	}

	var organizations int64
	err = database.WithinTenant(
		context.Background(),
		handle,
		database.TenantContext{OrganizationID: first.Organization.ID.String()},
		func(tx database.Tx) error {
			return tx.Session().Table("identity.organizations").
				Where("country_of_incorporation = ? AND business_registration_number = ? AND deleted_at IS NULL",
					country, number).
				Count(&organizations).Error
		},
	)
	if err != nil {
		t.Fatalf("count organizations: %v", err)
	}
	if organizations != 1 {
		t.Fatalf("expected the registration to be held by exactly one organization, got %d", organizations)
	}

	var reservations int64
	err = database.WithinTenant(
		context.Background(),
		handle,
		database.TenantContext{UserID: userID.String()},
		func(tx database.Tx) error {
			return tx.Session().Table("identity.idempotency_records").
				Where("endpoint = ? AND idempotency_key = ?",
					"POST /v1/organizations", "key-replay-1").
				Count(&reservations).Error
		},
	)
	if err != nil {
		t.Fatalf("count reservations: %v", err)
	}
	if reservations != 1 {
		t.Fatalf("expected one idempotency record, got %d", reservations)
	}
}

func TestSecondOrganizationForSameUserIsRefused(t *testing.T) {
	handle := store(t)
	_, subject := seedUser(t, handle)
	country, number, legalName := registryRow(t, handle, "entity_status = 'active' AND NOT sanctioned")

	creator := organizations(handle)

	if _, err := create(t, handle, creator, subject,
		registration(legalName, country, number, "key-first-org")); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := create(t, handle, creator, subject,
		registration("Another Company Ltd.", "TW", "TW-11111111", "key-second-org"))
	if !errors.Is(err, service.ErrOrganizationExists) {
		t.Fatalf("expected ErrOrganizationExists, got %v", err)
	}
}
