package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

func facilityService(handle *gorm.DB) *service.FacilityService {
	return service.NewFacilityService(
		handle,
		repository.NewFacilityRepository(),
		repository.NewOrganizationRepository(),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func seedVerifiedOrganization(
	t *testing.T,
	handle *gorm.DB,
	userID uuid.UUID,
	status domain.VerificationStatus,
) (uuid.UUID, string) {
	t.Helper()

	organizationID := uuid.New()
	registration := "PROBE-" + organizationID.String()[:13]

	err := handle.Exec(
		`INSERT INTO identity.organizations
		   (id, name, type, country_of_incorporation,
		    business_registration_number, verification_status)
		 VALUES (?, ?, 'manufacturer', 'TW', ?, ?::identity.verification_status)`,
		organizationID, "Probe Org "+organizationID.String()[:8], registration, string(status),
	).Error
	if err != nil {
		t.Fatalf("seed organization: %v", err)
	}

	t.Cleanup(func() {
		scoped := database.TenantContext{
			UserID:         userID.String(),
			OrganizationID: organizationID.String(),
		}
		cleanupErr := database.WithinTenant(context.Background(), handle, scoped,
			func(tx database.Tx) error {
				tx.Session().Exec(`DELETE FROM identity.facilities WHERE organization_id = ?`, organizationID)
				tx.Session().Exec(
					`DELETE FROM identity.idempotency_records WHERE organization_id = ? OR user_id = ?`,
					organizationID, userID,
				)
				return tx.Session().Exec(
					`UPDATE identity.organizations SET deleted_at = now() WHERE id = ?`, organizationID,
				).Error
			})
		if cleanupErr != nil {
			t.Errorf("clean organization %s: %v", organizationID, cleanupErr)
		}
	})

	return organizationID, registration
}

func seedFacilityRegistryRecord(
	t *testing.T,
	handle *gorm.DB,
	registration, reference, capacity, energy string,
) {
	t.Helper()

	err := handle.Exec(
		`INSERT INTO identity.facility_registry_records
		   (organization_registration_number, facility_reference,
		    attested_capacity, attested_energy_kwh)
		 VALUES (?, ?, ?::numeric, ?::numeric)`,
		registration, reference, capacity, energy,
	).Error
	if err != nil {
		t.Fatalf("seed facility registry record: %v", err)
	}

	t.Cleanup(func() {
		handle.Exec(
			`DELETE FROM identity.facility_registry_records
			 WHERE organization_registration_number = ?`, registration,
		)
	})
}

func declaration(reference, key string) service.FacilityDeclaration {
	return service.FacilityDeclaration{
		Name:              "Probe Site",
		Address:           "1 Probe Road, Hsinchu",
		CountryCode:       "TW",
		GridRegion:        "TW",
		Type:              domain.FacilityComponentFabrication,
		FacilityReference: reference,
		DeclaredCapacity:  "1000000.000000",
		DeclaredEnergyKwh: "2000000.000000",
		IdempotencyKey:    key,
		RequestBody:       []byte(reference + "|" + key),
	}
}

func managerOf(organizationID, userID uuid.UUID) service.Actor {
	return service.Actor{
		OrganizationID: organizationID,
		UserID:         userID,
		Role:           domain.RoleOwner,
	}
}

func TestMatchedFacilityUsesAttestedFiguresAndNoDiscount(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID, registration := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationVerified,
	)
	seedFacilityRegistryRecord(
		t, handle, registration, "TW-PRB-01", "18000000.000000", "31000000.000000",
	)

	facilities := facilityService(handle)

	registered, err := facilities.Create(
		context.Background(),
		managerOf(organizationID, userID),
		declaration("TW-PRB-01", "facility-key-matched"),
	)
	if err != nil {
		t.Fatalf("create facility: %v", err)
	}

	facility := registered.Facility

	if facility.VerificationStatus != domain.FacilityMatched {
		t.Fatalf("expected facility_matched, got %s", facility.VerificationStatus)
	}
	if facility.CeilingDiscountFactor != "1.00" {
		t.Fatalf("expected discount 1.00, got %s", facility.CeilingDiscountFactor)
	}
	if facility.AttestedCapacity == nil || *facility.AttestedCapacity != "18000000.000000" {
		t.Fatalf("expected attested capacity from the registry, got %v", facility.AttestedCapacity)
	}
	if facility.AttestedEnergyKwh == nil || *facility.AttestedEnergyKwh != "31000000.000000" {
		t.Fatalf("expected attested energy from the registry, got %v", facility.AttestedEnergyKwh)
	}
}

func TestVerifiedOrganizationWithUnmatchedFacilityTakesQuarterDiscount(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID, _ := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationVerified,
	)

	facilities := facilityService(handle)

	registered, err := facilities.Create(
		context.Background(),
		managerOf(organizationID, userID),
		declaration("", "facility-key-org-matched"),
	)
	if err != nil {
		t.Fatalf("create facility: %v", err)
	}

	if registered.Facility.VerificationStatus != domain.OrganizationMatched {
		t.Fatalf("expected organization_matched, got %s", registered.Facility.VerificationStatus)
	}
	if registered.Facility.CeilingDiscountFactor != "0.75" {
		t.Fatalf("expected discount 0.75, got %s", registered.Facility.CeilingDiscountFactor)
	}
	if registered.Facility.AttestedCapacity != nil {
		t.Fatalf("expected no attested capacity, got %v", *registered.Facility.AttestedCapacity)
	}
}

func TestUnverifiedOrganizationTakesHalfDiscount(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID, _ := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationUnverified,
	)

	facilities := facilityService(handle)

	registered, err := facilities.Create(
		context.Background(),
		managerOf(organizationID, userID),
		declaration("", "facility-key-self-declared"),
	)
	if err != nil {
		t.Fatalf("create facility: %v", err)
	}

	if registered.Facility.VerificationStatus != domain.SelfDeclared {
		t.Fatalf("expected self_declared, got %s", registered.Facility.VerificationStatus)
	}
	if registered.Facility.CeilingDiscountFactor != "0.50" {
		t.Fatalf("expected discount 0.50, got %s", registered.Facility.CeilingDiscountFactor)
	}
}

func TestReferenceBelongingToAnotherOrganizationDoesNotMatch(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID, _ := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationVerified,
	)

	strangerUser, _ := seedUser(t, handle)
	_, strangerRegistration := seedVerifiedOrganization(
		t, handle, strangerUser, domain.VerificationVerified,
	)
	seedFacilityRegistryRecord(
		t, handle, strangerRegistration, "TW-OTH-01", "99000000.000000", "99000000.000000",
	)

	facilities := facilityService(handle)

	registered, err := facilities.Create(
		context.Background(),
		managerOf(organizationID, userID),
		declaration("TW-OTH-01", "facility-key-foreign-reference"),
	)
	if err != nil {
		t.Fatalf("create facility: %v", err)
	}

	if registered.Facility.VerificationStatus == domain.FacilityMatched {
		t.Fatal("a reference registered to another organization must not match")
	}
	if registered.Facility.AttestedCapacity != nil {
		t.Fatalf("another organization's attested capacity leaked: %v", *registered.Facility.AttestedCapacity)
	}
}

func TestFacilityCreationReplaysOnTheSameIdempotencyKey(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID, _ := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationVerified,
	)

	facilities := facilityService(handle)
	actor := managerOf(organizationID, userID)

	first, err := facilities.Create(context.Background(), actor, declaration("", "facility-key-replay"))
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	second, err := facilities.Create(context.Background(), actor, declaration("", "facility-key-replay"))
	if err != nil {
		t.Fatalf("replayed create: %v", err)
	}

	if !second.Replayed {
		t.Fatal("expected the second call to be a replay")
	}
	if second.Facility.ID != first.Facility.ID {
		t.Fatalf("replay produced a different facility: %s then %s",
			first.Facility.ID, second.Facility.ID)
	}

	listed, err := facilities.List(context.Background(), actor)
	if err != nil {
		t.Fatalf("list facilities: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected exactly one facility after a replay, got %d", len(listed))
	}
}

func TestMemberCannotCreateAFacility(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID, _ := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationVerified,
	)

	facilities := facilityService(handle)

	_, err := facilities.Create(context.Background(), service.Actor{
		OrganizationID: organizationID,
		UserID:         userID,
		Role:           domain.RoleMember,
	}, declaration("", "facility-key-member"))

	if !errors.Is(err, service.ErrNotPermitted) {
		t.Fatalf("expected ErrNotPermitted, got %v", err)
	}
}

func TestAnotherOrganizationCannotReadYourFacility(t *testing.T) {
	handle := store(t)

	ownerUser, _ := seedUser(t, handle)
	ownerOrganization, _ := seedVerifiedOrganization(
		t, handle, ownerUser, domain.VerificationVerified,
	)

	strangerUser, _ := seedUser(t, handle)
	strangerOrganization, _ := seedVerifiedOrganization(
		t, handle, strangerUser, domain.VerificationVerified,
	)

	facilities := facilityService(handle)

	registered, err := facilities.Create(
		context.Background(),
		managerOf(ownerOrganization, ownerUser),
		declaration("", "facility-key-isolation"),
	)
	if err != nil {
		t.Fatalf("create facility: %v", err)
	}

	_, err = facilities.Get(
		context.Background(),
		managerOf(strangerOrganization, strangerUser),
		registered.Facility.ID,
	)
	if !errors.Is(err, service.ErrFacilityNotFound) {
		t.Fatalf("expected another organization to be refused, got %v", err)
	}

	listed, err := facilities.List(
		context.Background(),
		managerOf(strangerOrganization, strangerUser),
	)
	if err != nil {
		t.Fatalf("list as stranger: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("stranger saw %d facilities that are not theirs", len(listed))
	}
}

func TestFacilityRejectsUnknownGridRegionAndZeroFigures(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID, _ := seedVerifiedOrganization(
		t, handle, userID, domain.VerificationVerified,
	)

	facilities := facilityService(handle)
	actor := managerOf(organizationID, userID)

	unknownRegion := declaration("", "facility-key-region")
	unknownRegion.GridRegion = "MARS-1"
	if _, err := facilities.Create(context.Background(), actor, unknownRegion); err == nil {
		t.Fatal("expected an unknown grid region to be rejected")
	}

	zeroCapacity := declaration("", "facility-key-zero")
	zeroCapacity.DeclaredCapacity = "0.000000"
	if _, err := facilities.Create(context.Background(), actor, zeroCapacity); err == nil {
		t.Fatal("expected a zero declared capacity to be rejected")
	}
}
