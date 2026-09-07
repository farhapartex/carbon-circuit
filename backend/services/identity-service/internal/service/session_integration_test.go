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
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

func sessionRegistry(handle *gorm.DB) *service.SessionRegistry {
	return service.NewSessionRegistry(
		handle,
		repository.NewSessionRepository(),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func seedSessionUser(t *testing.T, handle *gorm.DB) uuid.UUID {
	t.Helper()

	userID, _ := seedUser(t, handle)

	t.Cleanup(func() {
		scoped := database.TenantContext{UserID: userID.String()}
		err := database.WithinTenant(context.Background(), handle, scoped,
			func(tx database.Tx) error {
				return tx.Session().Exec(
					`DELETE FROM identity.sessions WHERE user_id = ?`, userID,
				).Error
			})
		if err != nil {
			t.Errorf("clean sessions for %s: %v", userID, err)
		}
	})

	return userID
}

func sighting(userID uuid.UUID, sessionID, agent string) service.SessionSighting {
	return service.SessionSighting{
		UserID:         userID,
		Auth0SessionID: sessionID,
		UserAgent:      agent,
		IPAddress:      "203.0.113.7",
	}
}

func TestRecordingTheSameSessionTwiceKeepsOneRow(t *testing.T) {
	handle := store(t)
	userID := seedSessionUser(t, handle)
	registry := sessionRegistry(handle)

	first := sighting(userID, "sess-laptop", "Firefox on macOS")
	if err := registry.Record(context.Background(), first); err != nil {
		t.Fatalf("record session: %v", err)
	}

	again := sighting(userID, "sess-laptop", "Firefox on macOS")
	if err := registry.Record(context.Background(), again); err != nil {
		t.Fatalf("record session again: %v", err)
	}

	listed, err := registry.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected the same session id to upsert, got %d rows", len(listed))
	}
}

func TestEachDeviceGetsItsOwnSession(t *testing.T) {
	handle := store(t)
	userID := seedSessionUser(t, handle)
	registry := sessionRegistry(handle)

	for _, device := range []struct{ id, agent string }{
		{"sess-laptop", "Firefox on macOS"},
		{"sess-phone", "Safari on iOS"},
	} {
		if err := registry.Record(
			context.Background(), sighting(userID, device.id, device.agent),
		); err != nil {
			t.Fatalf("record %s: %v", device.id, err)
		}
	}

	listed, err := registry.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected two devices, got %d", len(listed))
	}
}

func TestRevokingOneDeviceLeavesTheOtherListed(t *testing.T) {
	handle := store(t)
	userID := seedSessionUser(t, handle)
	registry := sessionRegistry(handle)

	for _, id := range []string{"sess-laptop", "sess-phone"} {
		if err := registry.Record(
			context.Background(), sighting(userID, id, "device"),
		); err != nil {
			t.Fatalf("record %s: %v", id, err)
		}
	}

	if err := registry.Revoke(context.Background(), userID, "sess-laptop"); err != nil {
		t.Fatalf("revoke session: %v", err)
	}

	listed, err := registry.List(context.Background(), userID)
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one remaining session, got %d", len(listed))
	}
	if listed[0].Auth0SessionID != "sess-phone" {
		t.Fatalf("the wrong device survived: %s", listed[0].Auth0SessionID)
	}
}

func TestRevokingAnAlreadyRevokedSessionIsRefused(t *testing.T) {
	handle := store(t)
	userID := seedSessionUser(t, handle)
	registry := sessionRegistry(handle)

	if err := registry.Record(
		context.Background(), sighting(userID, "sess-laptop", "device"),
	); err != nil {
		t.Fatalf("record session: %v", err)
	}

	if err := registry.Revoke(context.Background(), userID, "sess-laptop"); err != nil {
		t.Fatalf("first revoke: %v", err)
	}

	err := registry.Revoke(context.Background(), userID, "sess-laptop")
	if !errors.Is(err, service.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound on a second revoke, got %v", err)
	}
}

func TestOneUserNeverSeesAnotherUsersSessions(t *testing.T) {
	handle := store(t)
	owner := seedSessionUser(t, handle)
	stranger := seedSessionUser(t, handle)
	registry := sessionRegistry(handle)

	if err := registry.Record(
		context.Background(), sighting(owner, "sess-owner", "Owner laptop"),
	); err != nil {
		t.Fatalf("record owner session: %v", err)
	}

	listed, err := registry.List(context.Background(), stranger)
	if err != nil {
		t.Fatalf("list as stranger: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("a stranger saw %d sessions that are not theirs", len(listed))
	}

	if err := registry.Revoke(
		context.Background(), stranger, "sess-owner",
	); !errors.Is(err, service.ErrSessionNotFound) {
		t.Fatalf("a stranger must not revoke another user's session, got %v", err)
	}
}
