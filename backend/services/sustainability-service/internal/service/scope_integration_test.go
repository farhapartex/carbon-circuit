package service_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/repository"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
)

func store(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run sustainability integration tests")
	}

	opened, err := database.Open(context.Background(), database.Options{
		DSN:             dsn,
		Schema:          "sustainability",
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

func quiet() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func seedClaim(t *testing.T, handle *gorm.DB, organizationID uuid.UUID) uuid.UUID {
	t.Helper()

	claimID := uuid.New()
	userID := uuid.New()

	scoped := database.TenantContext{
		UserID:         userID.String(),
		OrganizationID: organizationID.String(),
	}

	err := database.WithinTenant(context.Background(), handle, scoped, func(tx database.Tx) error {
		return tx.Session().Exec(`
			INSERT INTO sustainability.claims
			  (id, organization_id, submitted_by_user_id, facility_id, facility_name,
			   activity_type, vintage_year, period_start, period_end, declared_figures,
			   requested_amount, computed_ceiling, capacity_basis, capacity_source,
			   discount_factor, reference_factor_id, reference_factor_value,
			   status, priority, exclusivity_attested_at, exclusivity_attested_by)
			VALUES (?, ?, ?, ?, 'Probe Facility', 'renewable_energy', 2031,
			        '2031-01-01', '2031-12-31', '{}'::jsonb,
			        100, 500, 1000, 'attested', 1.00,
			        (SELECT id FROM sustainability.reference_factors LIMIT 1), 0.494,
			        'human_review', 'normal', now(), ?)`,
			claimID, organizationID, userID, uuid.New(), userID).Error
	})
	if err != nil {
		t.Fatalf("seed claim: %v", err)
	}

	t.Cleanup(func() {
		database.WithinTenant(context.Background(), handle, scoped, func(tx database.Tx) error {
			return tx.Session().Exec(
				`DELETE FROM sustainability.claims WHERE id = ?`, claimID).Error
		})
	})

	return claimID
}

func TestAVerifiersOwnClaimListStaysScopedToTheirOrganization(t *testing.T) {
	handle := store(t)

	stranger := uuid.New()
	verifierOrganization := uuid.New()

	seedClaim(t, handle, stranger)
	seedClaim(t, handle, stranger)
	ownClaim := seedClaim(t, handle, verifierOrganization)

	claims := service.NewClaimService(
		handle, repository.NewClaimRepository(), repository.NewReferenceRepository(),
		nil, nil, quiet(),
	)

	verifiers := service.NewVerifierService(handle, repository.NewClaimRepository(), quiet())

	reviewer := uuid.New()

	tenantPage, err := claims.List(context.Background(), service.Actor{
		OrganizationID:    verifierOrganization,
		UserID:            reviewer,
		OrganizationState: "active",
	}, "", "", 50)
	if err != nil {
		t.Fatalf("tenant list: %v", err)
	}

	if len(tenantPage.Claims) != 1 {
		t.Fatalf("a verifier browsing their own claims must see only their organization's, got %d",
			len(tenantPage.Claims))
	}
	if tenantPage.Claims[0].ID != ownClaim {
		t.Fatal("the tenant list returned a claim from another organization")
	}

	queue, err := verifiers.Queue(context.Background(), service.Verifier{
		UserID:       reviewer,
		Name:         "Probe Verifier",
		PlatformRole: service.VerifierRole,
	}, "", 50)
	if err != nil {
		t.Fatalf("verifier queue: %v", err)
	}

	if len(queue.Claims) < 3 {
		t.Fatalf("the verifier queue must reach every organization, got %d", len(queue.Claims))
	}
}

func TestTheTenantPathCannotCarryAPlatformRole(t *testing.T) {
	actor := service.Actor{OrganizationID: uuid.New(), UserID: uuid.New()}

	settings, err := database.SettingsFor(service.TenancyOf(actor))
	if err != nil {
		t.Fatalf("settings: %v", err)
	}

	if _, present := settings["app.platform_role"]; present {
		t.Fatal("an ordinary claim request must never carry a platform role into the transaction")
	}
}

func TestAnExcludedVerifierSeesNothingFromThatOrganization(t *testing.T) {
	handle := store(t)

	subject := uuid.New()
	seedClaim(t, handle, subject)

	reviewer := uuid.New()
	verifiers := service.NewVerifierService(handle, repository.NewClaimRepository(), quiet())
	who := service.Verifier{UserID: reviewer, Name: "Probe", PlatformRole: service.VerifierRole}

	before, err := verifiers.Queue(context.Background(), who, "", 100)
	if err != nil {
		t.Fatalf("queue before: %v", err)
	}

	declareExclusion(t, handle, reviewer, subject)

	after, err := verifiers.Queue(context.Background(), who, "", 100)
	if err != nil {
		t.Fatalf("queue after: %v", err)
	}

	if len(after.Claims) >= len(before.Claims) {
		t.Fatalf("declaring a relationship must remove that organization's claims: %d then %d",
			len(before.Claims), len(after.Claims))
	}

	for _, claim := range after.Claims {
		if claim.OrganizationID == subject {
			t.Fatal("an excluded organization's claim still reached the queue")
		}
	}
}

func declareExclusion(t *testing.T, handle *gorm.DB, reviewer, organizationID uuid.UUID) {
	t.Helper()

	err := handle.Exec(`
		INSERT INTO sustainability.verifier_exclusions
		  (verifier_user_id, organization_id, declared_by, reason)
		VALUES (?, ?, ?, 'probe relationship')`,
		reviewer, organizationID, reviewer).Error
	if err != nil {
		t.Fatalf("declare exclusion: %v", err)
	}

	t.Cleanup(func() {
		handle.Exec(`DELETE FROM sustainability.verifier_exclusions WHERE verifier_user_id = ?`, reviewer)
	})
}

var _ = domain.HumanReview
