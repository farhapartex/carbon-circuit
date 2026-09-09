package database_test

import (
	"strings"
	"testing"

	"github.com/carboncircuit/backend/internal/database"
)

func TestAPlatformRoleWithoutAnIdentityIsRefused(t *testing.T) {
	tenant := database.TenantContext{PlatformRole: "verifier"}

	if _, err := database.SettingsFor(tenant); err == nil {
		t.Fatal("a platform role with no user id makes every exclusion comparison null, and must be refused")
	}
}

func TestAPlatformRoleWithAnIdentityIsCarried(t *testing.T) {
	tenant := database.TenantContext{
		UserID:       "3f0f2f6c-0c4a-4f8f-9a3b-1a2b3c4d5e6f",
		PlatformRole: "verifier",
	}

	settings, err := database.SettingsFor(tenant)
	if err != nil {
		t.Fatalf("settings: %v", err)
	}

	if settings["app.platform_role"] != "verifier" {
		t.Fatalf("expected the platform role to be carried, got %q", settings["app.platform_role"])
	}
	if settings["app.user_id"] == "" {
		t.Fatal("expected the user id to be carried alongside the role")
	}
}

func TestAnOrdinaryTenantCarriesNoPlatformRole(t *testing.T) {
	tenant := database.TenantContext{
		UserID:         "3f0f2f6c-0c4a-4f8f-9a3b-1a2b3c4d5e6f",
		OrganizationID: "8d1c2b3a-4e5f-6789-0abc-def012345678",
	}

	settings, err := database.SettingsFor(tenant)
	if err != nil {
		t.Fatalf("settings: %v", err)
	}

	if _, present := settings["app.platform_role"]; present {
		t.Fatal("an ordinary tenant request must never set a platform role")
	}
}

func TestANonUuidIdentityIsRefused(t *testing.T) {
	tenant := database.TenantContext{UserID: "not-a-uuid"}

	_, err := database.SettingsFor(tenant)
	if err == nil || !strings.Contains(err.Error(), "uuid") {
		t.Fatalf("expected a uuid complaint, got %v", err)
	}
}
