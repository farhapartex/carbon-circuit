package outbox_test

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
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/internal/outbox"
)

type recordingDispatcher struct {
	mutex    sync.Mutex
	captured []kafka.Message
	fail     error
	entered  chan struct{}
	release  chan struct{}
}

func (d *recordingDispatcher) Publish(_ context.Context, messages ...kafka.Message) error {
	if d.entered != nil {
		d.entered <- struct{}{}
		<-d.release
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.fail != nil {
		return d.fail
	}

	d.captured = append(d.captured, messages...)
	return nil
}

func (d *recordingDispatcher) count() int {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return len(d.captured)
}

func store(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run outbox integration tests")
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

	opened.Exec(`DELETE FROM identity.outbox_events`)
	t.Cleanup(func() { opened.Exec(`DELETE FROM identity.outbox_events`) })

	return opened
}

func appendEvent(t *testing.T, handle *gorm.DB, aggregateID uuid.UUID) uuid.UUID {
	t.Helper()

	var eventID uuid.UUID
	err := database.WithinTenant(
		context.Background(),
		handle,
		database.TenantContext{OrganizationID: aggregateID.String()},
		func(tx database.Tx) error {
			var appendErr error
			eventID, appendErr = outbox.Append(tx, outbox.Envelope{
				AggregateType: "organization",
				AggregateID:   aggregateID,
				EventType:     "organization.verified",
				Payload:       map[string]string{"organization_id": aggregateID.String()},
			})
			return appendErr
		},
	)
	if err != nil {
		t.Fatalf("append event: %v", err)
	}

	return eventID
}

func publisher(handle *gorm.DB, dispatch outbox.Dispatcher) *outbox.Publisher {
	return outbox.NewPublisher(outbox.PublisherOptions{
		Database:  handle,
		Dispatch:  dispatch,
		Logger:    database.DiscardLogger(),
		Interval:  time.Second,
		BatchSize: 50,
	})
}

func unpublished(t *testing.T, handle *gorm.DB) int {
	t.Helper()

	var count int64
	if err := handle.Table("identity.outbox_events").
		Where("published_at IS NULL").Count(&count).Error; err != nil {
		t.Fatalf("count unpublished: %v", err)
	}
	return int(count)
}

func TestAppendOutsideTransactionIsRefused(t *testing.T) {
	_, err := outbox.Append(database.Tx{}, outbox.Envelope{
		AggregateType: "organization",
		AggregateID:   uuid.New(),
		EventType:     "organization.verified",
	})
	if !errors.Is(err, database.ErrNotInTransaction) {
		t.Fatalf("expected ErrNotInTransaction, got %v", err)
	}
}

func TestDrainPublishesAndMarks(t *testing.T) {
	handle := store(t)
	organizationID := uuid.New()
	appendEvent(t, handle, organizationID)

	dispatch := &recordingDispatcher{}
	drained, err := publisher(handle, dispatch).DrainOnce(context.Background())
	if err != nil {
		t.Fatalf("drain: %v", err)
	}

	if drained != 1 || dispatch.count() != 1 {
		t.Fatalf("expected one event drained and dispatched, got %d and %d", drained, dispatch.count())
	}

	message := dispatch.captured[0]
	if message.Topic != "organization.verified" {
		t.Fatalf("expected the event type as topic, got %q", message.Topic)
	}
	if string(message.Key) != organizationID.String() {
		t.Fatal("expected the aggregate id as the partition key")
	}
	if message.Headers["event_id"] == "" {
		t.Fatal("expected an event_id header for consumer inbox de-duplication")
	}

	if remaining := unpublished(t, handle); remaining != 0 {
		t.Fatalf("expected no unpublished events, got %d", remaining)
	}
}

func TestDrainIsIdempotent(t *testing.T) {
	handle := store(t)
	appendEvent(t, handle, uuid.New())

	dispatch := &recordingDispatcher{}
	sender := publisher(handle, dispatch)

	if _, err := sender.DrainOnce(context.Background()); err != nil {
		t.Fatalf("first drain: %v", err)
	}

	drained, err := sender.DrainOnce(context.Background())
	if err != nil {
		t.Fatalf("second drain: %v", err)
	}

	if drained != 0 || dispatch.count() != 1 {
		t.Fatalf("expected the second drain to publish nothing, got %d dispatched in total", dispatch.count())
	}
}

func TestFailedDispatchLeavesEventUnpublished(t *testing.T) {
	handle := store(t)
	appendEvent(t, handle, uuid.New())

	dispatch := &recordingDispatcher{fail: errors.New("broker unreachable")}
	if _, err := publisher(handle, dispatch).DrainOnce(context.Background()); err == nil {
		t.Fatal("expected the drain to report the dispatch failure")
	}

	if remaining := unpublished(t, handle); remaining != 1 {
		t.Fatalf("expected the event to remain unpublished for retry, got %d", remaining)
	}
}

func TestConcurrentPublishersDoNotDoublePublish(t *testing.T) {
	handle := store(t)
	appendEvent(t, handle, uuid.New())

	blocking := &recordingDispatcher{
		entered: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	idle := &recordingDispatcher{}

	var group sync.WaitGroup
	group.Add(1)
	go func() {
		defer group.Done()
		if _, err := publisher(handle, blocking).DrainOnce(context.Background()); err != nil {
			t.Errorf("blocking drain: %v", err)
		}
	}()

	<-blocking.entered

	drained, err := publisher(handle, idle).DrainOnce(context.Background())
	if err != nil {
		t.Fatalf("competing drain: %v", err)
	}

	close(blocking.release)
	group.Wait()

	if drained != 0 || idle.count() != 0 {
		t.Fatalf("a competing publisher must skip locked rows, but drained %d", drained)
	}
	if blocking.count() != 1 {
		t.Fatalf("expected the holder to publish exactly once, got %d", blocking.count())
	}
	if remaining := unpublished(t, handle); remaining != 0 {
		t.Fatalf("expected the event published exactly once, %d remain", remaining)
	}
}
