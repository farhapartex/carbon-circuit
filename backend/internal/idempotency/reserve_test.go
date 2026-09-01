package idempotency_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/idempotency"
)

const endpoint = "POST /v1/organizations"

func store(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run idempotency integration tests")
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

func seedUser(t *testing.T, handle *gorm.DB) uuid.UUID {
	t.Helper()

	id := uuid.New()
	subject := "probe|" + id.String()

	err := handle.Exec(
		`INSERT INTO identity.users (auth0_subject, email, name, email_verified)
		 VALUES (?, ?, 'Probe User', true)`,
		subject, id.String()+"@probe.test",
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

	created, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("parse seeded user id: %v", err)
	}

	t.Cleanup(func() {
		handle.Exec(`DELETE FROM identity.idempotency_records WHERE user_id = ?`, created)
		handle.Exec(`DELETE FROM identity.users WHERE id = ?`, created)
	})

	return created
}

func request(userID uuid.UUID, key string, body string) idempotency.Request {
	return idempotency.Request{
		Scope:    idempotency.ForUser(userID),
		Endpoint: endpoint,
		Key:      key,
		Body:     []byte(body),
	}
}

func reserve(
	handle *gorm.DB,
	userID uuid.UUID,
	req idempotency.Request,
	then func(tx database.Tx, reservation idempotency.Reservation) error,
) (idempotency.Reservation, error) {
	var reservation idempotency.Reservation

	err := database.WithinTenant(
		context.Background(),
		handle,
		database.TenantContext{UserID: userID.String()},
		func(tx database.Tx) error {
			var reserveErr error
			reservation, reserveErr = idempotency.Reserve(tx, req)
			if reserveErr != nil {
				return reserveErr
			}
			if then == nil {
				return nil
			}
			return then(tx, reservation)
		},
	)

	return reservation, err
}

func TestFirstReservationSucceeds(t *testing.T) {
	handle := store(t)
	userID := seedUser(t, handle)

	reservation, err := reserve(handle, userID, request(userID, "key-1", `{"name":"Acme"}`), nil)
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if reservation.IsReplay() {
		t.Fatal("a first reservation must not be a replay")
	}
	if reservation.RecordID == uuid.Nil {
		t.Fatal("expected a record id")
	}
}

func TestCompletedKeyReplaysStoredResponse(t *testing.T) {
	handle := store(t)
	userID := seedUser(t, handle)
	req := request(userID, "key-2", `{"name":"Acme"}`)
	resourceID := uuid.New()

	_, err := reserve(handle, userID, req, func(tx database.Tx, r idempotency.Reservation) error {
		return idempotency.Complete(tx, r.RecordID, idempotency.Response{
			Status:     201,
			Body:       []byte(`{"id":"created"}`),
			ResourceID: &resourceID,
		})
	})
	if err != nil {
		t.Fatalf("first attempt: %v", err)
	}

	replayed, err := reserve(handle, userID, req, nil)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}

	if !replayed.IsReplay() {
		t.Fatal("expected the retry to replay the stored response")
	}
	if replayed.Replay.Status != 201 {
		t.Fatalf("expected status 201, got %d", replayed.Replay.Status)
	}
	if string(replayed.Replay.Body) != `{"id":"created"}` {
		t.Fatalf("expected the stored body, got %s", replayed.Replay.Body)
	}
	if replayed.Replay.ResourceID == nil || *replayed.Replay.ResourceID != resourceID {
		t.Fatal("expected the stored resource id to be replayed")
	}
}

func TestSameKeyWithDifferentBodyIsRejected(t *testing.T) {
	handle := store(t)
	userID := seedUser(t, handle)

	_, err := reserve(handle, userID, request(userID, "key-3", `{"name":"Acme"}`),
		func(tx database.Tx, r idempotency.Reservation) error {
			return idempotency.Complete(tx, r.RecordID, idempotency.Response{Status: 201})
		})
	if err != nil {
		t.Fatalf("first attempt: %v", err)
	}

	_, err = reserve(handle, userID, request(userID, "key-3", `{"name":"Different"}`), nil)
	if !errors.Is(err, idempotency.ErrKeyReused) {
		t.Fatalf("expected ErrKeyReused, got %v", err)
	}
}

func TestFailedKeyBecomesReusable(t *testing.T) {
	handle := store(t)
	userID := seedUser(t, handle)
	req := request(userID, "key-4", `{"name":"Acme"}`)

	_, err := reserve(handle, userID, req, func(tx database.Tx, r idempotency.Reservation) error {
		return idempotency.Fail(tx, r.RecordID)
	})
	if err != nil {
		t.Fatalf("first attempt: %v", err)
	}

	retried, err := reserve(handle, userID, req, nil)
	if err != nil {
		t.Fatalf("expected a failed key to be reusable, got %v", err)
	}
	if retried.IsReplay() {
		t.Fatal("a retaken key must execute rather than replay")
	}
}

func TestConcurrentDuplicateLosesTheRace(t *testing.T) {
	handle := store(t)
	userID := seedUser(t, handle)
	req := request(userID, "key-5", `{"name":"Acme"}`)

	release := make(chan struct{})
	var group sync.WaitGroup
	outcomes := make([]error, 2)

	for i := range outcomes {
		group.Add(1)
		go func(slot int) {
			defer group.Done()
			outcomes[slot] = database.WithinTenant(
				context.Background(),
				handle,
				database.TenantContext{UserID: userID.String()},
				func(tx database.Tx) error {
					if _, err := idempotency.Reserve(tx, req); err != nil {
						return err
					}
					<-release
					return nil
				},
			)
		}(i)
	}

	time.Sleep(300 * time.Millisecond)
	close(release)
	group.Wait()

	var succeeded, blocked int
	for _, err := range outcomes {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, idempotency.ErrInProgress):
			blocked++
		default:
			t.Fatalf("unexpected outcome: %v", err)
		}
	}

	if succeeded != 1 || blocked != 1 {
		t.Fatalf("expected exactly one winner and one blocked, got %d and %d", succeeded, blocked)
	}
}

func TestReserveOutsideTransactionIsRefused(t *testing.T) {
	if _, err := idempotency.Reserve(database.Tx{}, request(uuid.New(), "key-6", "{}")); !errors.Is(err, database.ErrNotInTransaction) {
		t.Fatalf("expected ErrNotInTransaction, got %v", err)
	}
}
