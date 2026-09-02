package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/provenance-service/internal/domain"
	"github.com/carboncircuit/backend/services/provenance-service/internal/repository"
	"github.com/carboncircuit/backend/services/provenance-service/internal/score"
	"github.com/carboncircuit/backend/services/provenance-service/internal/service"
)

type stubFacilities struct {
	name    string
	missing bool
}

func (s stubFacilities) Facility(
	_ context.Context,
	facilityID uuid.UUID,
) (service.Facility, error) {
	if s.missing {
		return service.Facility{}, service.ErrFacilityUnknown
	}
	return service.Facility{ID: facilityID, Name: s.name}, nil
}

func store(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run provenance integration tests")
	}

	opened, err := database.Open(context.Background(), database.Options{
		DSN:             dsn,
		Schema:          "provenance",
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

func batchService(handle *gorm.DB, facilities service.FacilityResolver) *service.BatchService {
	return service.NewBatchService(
		handle,
		repository.NewBatchRepository(),
		repository.NewCheckpointRepository(),
		facilities,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func actor(t *testing.T, handle *gorm.DB, name string) service.Actor {
	t.Helper()

	organizationID := uuid.New()
	userID := uuid.New()

	t.Cleanup(func() {
		scoped := database.TenantContext{
			UserID:         userID.String(),
			OrganizationID: organizationID.String(),
		}
		err := database.WithinTenant(context.Background(), handle, scoped,
			func(tx database.Tx) error {
				tx.Session().Exec(`DELETE FROM provenance.checkpoints WHERE organization_id = ?`, organizationID)
				tx.Session().Exec(`DELETE FROM provenance.batch_parents WHERE organization_id = ?`, organizationID)
				tx.Session().Exec(`DELETE FROM provenance.idempotency_records WHERE organization_id = ?`, organizationID)
				tx.Session().Exec(`DELETE FROM provenance.outbox_events WHERE aggregate_id IN (SELECT id FROM provenance.batches WHERE organization_id = ?)`, organizationID)
				return tx.Session().Exec(`DELETE FROM provenance.batches WHERE organization_id = ?`, organizationID).Error
			})
		if err != nil {
			t.Errorf("clean organization %s: %v", organizationID, err)
		}
	})

	return service.Actor{
		OrganizationID:    organizationID,
		UserID:            userID,
		OrganizationName:  name,
		OrganizationType:  "manufacturer",
		OrganizationState: "active",
	}
}

func declaration(key string, parents ...string) service.BatchDeclaration {
	return service.BatchDeclaration{
		OriginatingFacilityID: uuid.New(),
		ProductCategory:       domain.CategoryElectronics,
		ComponentType:         "300mm silicon wafer",
		LotNumber:             "WL-1",
		Quantity:              "5000.000000",
		Unit:                  "wafers",
		ProducedAt:            time.Now().UTC().Add(-72 * time.Hour),
		ParentReferences:      parents,
		IdempotencyKey:        key,
		RequestBody:           []byte(key),
	}
}

func checkpointEntry(kind domain.CheckpointType, key string) service.CheckpointEntry {
	entry := service.CheckpointEntry{
		Type:           kind,
		LocationLabel:  "Hsinchu Fab",
		CountryCode:    "TW",
		OccurredAt:     time.Now().UTC().Add(-2 * time.Hour),
		IdempotencyKey: key,
		RequestBody:    []byte(key),
	}
	if kind != domain.ProductionComplete {
		entry.ShippingMethod = "sea_freight_container"
	}
	return entry
}

func componentsOf(t *testing.T, batch domain.Batch) []score.Component {
	t.Helper()

	var components []score.Component
	if err := json.Unmarshal(batch.ScoreComponents, &components); err != nil {
		t.Fatalf("decode score components: %v", err)
	}
	return components
}

func TestNewBatchGetsARandomReferenceAndScoresZero(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Formosa Precision")
	batches := batchService(handle, stubFacilities{name: "Hsinchu Fab TW-01"})

	created, err := batches.Create(context.Background(), owner, declaration("batch-key-1"))
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	if len(created.Batch.PublicReference) != domain.PublicReferenceLength {
		t.Fatalf("expected a %d character reference, got %q",
			domain.PublicReferenceLength, created.Batch.PublicReference)
	}
	if created.Batch.PublicReference == created.Batch.ID.String() {
		t.Fatal("the public reference must never be the internal identifier")
	}
	if created.Batch.ProvenanceScore != 0 {
		t.Fatalf("a batch with no checkpoints scores 0, got %d", created.Batch.ProvenanceScore)
	}
	if len(componentsOf(t, created.Batch)) != 5 {
		t.Fatal("expected all five score components to be stored")
	}
}

func TestLoggingCheckpointsRaisesTheStoredScore(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Formosa Precision")
	batches := batchService(handle, stubFacilities{name: "Hsinchu Fab TW-01"})

	created, err := batches.Create(context.Background(), owner, declaration("batch-key-2"))
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	logged, err := batches.LogCheckpoint(
		context.Background(), owner, created.Batch.ID,
		checkpointEntry(domain.ProductionComplete, "checkpoint-key-1"),
	)
	if err != nil {
		t.Fatalf("log checkpoint: %v", err)
	}

	if logged.Checkpoint.ReportedByOrganizationName != "Formosa Precision" {
		t.Fatalf("expected the reporter name from the token, got %q",
			logged.Checkpoint.ReportedByOrganizationName)
	}

	view, err := batches.Get(context.Background(), owner, created.Batch.ID)
	if err != nil {
		t.Fatalf("get batch: %v", err)
	}

	if view.Batch.CheckpointCount != 1 {
		t.Fatalf("expected one checkpoint counted, got %d", view.Batch.CheckpointCount)
	}
	if view.Batch.ProvenanceScore != 38 {
		t.Fatalf("expected 8 completeness + 15 chain depth + 15 timeliness = 38, got %d",
			view.Batch.ProvenanceScore)
	}
}

func TestOnlyTheOwnerMayLogCheckpoints(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Owner Org")
	stranger := actor(t, handle, "Stranger Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	created, err := batches.Create(context.Background(), owner, declaration("batch-key-3"))
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	_, err = batches.LogCheckpoint(
		context.Background(), stranger, created.Batch.ID,
		checkpointEntry(domain.ProductionComplete, "checkpoint-key-stranger"),
	)
	if err == nil {
		t.Fatal("a stranger must not be able to log a checkpoint on someone else's batch")
	}
	if !errors.Is(err, service.ErrBatchNotFound) && !errors.Is(err, service.ErrNotBatchOwner) {
		t.Fatalf("expected the write to be refused, got %v", err)
	}
}

func TestCreatingABatchReplaysOnTheSameKey(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Owner Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	first, err := batches.Create(context.Background(), owner, declaration("batch-key-replay"))
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	second, err := batches.Create(context.Background(), owner, declaration("batch-key-replay"))
	if err != nil {
		t.Fatalf("replayed create: %v", err)
	}

	if !second.Replayed {
		t.Fatal("expected the second call to replay")
	}
	if second.Batch.ID != first.Batch.ID {
		t.Fatalf("replay produced a different batch: %s then %s", first.Batch.ID, second.Batch.ID)
	}

	page, err := batches.List(context.Background(), owner, "", 25)
	if err != nil {
		t.Fatalf("list batches: %v", err)
	}
	if len(page.Batches) != 1 {
		t.Fatalf("expected exactly one batch after a replay, got %d", len(page.Batches))
	}
}

func TestDeclaredParentResolvesAndScoresChainDepth(t *testing.T) {
	handle := store(t)
	supplier := actor(t, handle, "Supplier Org")
	assembler := actor(t, handle, "Assembler Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	upstream, err := batches.Create(context.Background(), supplier, declaration("batch-key-parent"))
	if err != nil {
		t.Fatalf("create parent batch: %v", err)
	}

	child, err := batches.Create(
		context.Background(), assembler,
		declaration("batch-key-child", upstream.Batch.PublicReference),
	)
	if err != nil {
		t.Fatalf("create child batch: %v", err)
	}

	if len(child.Parents) != 1 {
		t.Fatalf("expected one declared parent, got %d", len(child.Parents))
	}
	if !child.Parents[0].Resolved {
		t.Fatal("a reference matching a registered batch must resolve")
	}
	if child.Parents[0].Batch == nil {
		t.Fatal("a resolved parent must carry its disclosable fields")
	}
}

func TestUnresolvedParentIsRecordedAndCostsChainDepth(t *testing.T) {
	handle := store(t)
	assembler := actor(t, handle, "Assembler Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	created, err := batches.Create(
		context.Background(), assembler,
		declaration("batch-key-unresolved", "ZZZZZZZZZZZZZZZZZZZZZZ"),
	)
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	if len(created.Parents) != 1 {
		t.Fatalf("expected the declaration to be recorded, got %d", len(created.Parents))
	}
	if created.Parents[0].Resolved {
		t.Fatal("a reference matching nothing must not resolve")
	}

	if _, err := batches.LogCheckpoint(
		context.Background(), assembler, created.Batch.ID,
		checkpointEntry(domain.ProductionComplete, "checkpoint-key-unresolved"),
	); err != nil {
		t.Fatalf("log checkpoint: %v", err)
	}

	view, err := batches.Get(context.Background(), assembler, created.Batch.ID)
	if err != nil {
		t.Fatalf("get batch: %v", err)
	}

	for _, component := range componentsOf(t, view.Batch) {
		if component.Label == "Chain depth resolution" && component.Earned != 0 {
			t.Fatalf("an unresolved declaration must cost the chain depth points, got %d", component.Earned)
		}
	}
}

func TestComponentViewIsVisibleOneLevelUpAndNowhereElse(t *testing.T) {
	handle := store(t)
	supplier := actor(t, handle, "Supplier Org")
	assembler := actor(t, handle, "Assembler Org")
	outsider := actor(t, handle, "Outsider Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	upstream, err := batches.Create(context.Background(), supplier, declaration("batch-key-up"))
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}

	if _, err := batches.LogCheckpoint(
		context.Background(), supplier, upstream.Batch.ID,
		checkpointEntry(domain.ProductionComplete, "checkpoint-key-up"),
	); err != nil {
		t.Fatalf("log checkpoint on parent: %v", err)
	}

	child, err := batches.Create(
		context.Background(), assembler,
		declaration("batch-key-down", upstream.Batch.PublicReference),
	)
	if err != nil {
		t.Fatalf("create child: %v", err)
	}

	component, err := batches.Component(
		context.Background(), assembler, child.Batch.ID, upstream.Batch.ID,
	)
	if err != nil {
		t.Fatalf("assembler must see its declared parent: %v", err)
	}
	if len(component.Checkpoints) != 1 {
		t.Fatalf("expected the parent's checkpoint history, got %d", len(component.Checkpoints))
	}

	if _, err := batches.Component(
		context.Background(), outsider, child.Batch.ID, upstream.Batch.ID,
	); !errors.Is(err, service.ErrBatchNotFound) {
		t.Fatalf("an unrelated organization must be refused, got %v", err)
	}

	if _, err := batches.Get(context.Background(), assembler, upstream.Batch.ID); !errors.Is(err, service.ErrBatchNotFound) {
		t.Fatalf("the owner view of a parent must stay closed to the child's owner, got %v", err)
	}
}

func TestCheckpointCannotPrecedeProductionOrSitInTheFuture(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Owner Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	created, err := batches.Create(context.Background(), owner, declaration("batch-key-time"))
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	early := checkpointEntry(domain.ProductionComplete, "checkpoint-key-early")
	early.OccurredAt = created.Batch.ProducedAt.Add(-time.Hour)
	if _, err := batches.LogCheckpoint(context.Background(), owner, created.Batch.ID, early); !errors.Is(err, service.ErrCheckpointBeforeProduction) {
		t.Fatalf("expected ErrCheckpointBeforeProduction, got %v", err)
	}

	future := checkpointEntry(domain.ProductionComplete, "checkpoint-key-future")
	future.OccurredAt = time.Now().UTC().Add(time.Hour)
	if _, err := batches.LogCheckpoint(context.Background(), owner, created.Batch.ID, future); !errors.Is(err, service.ErrCheckpointInFuture) {
		t.Fatalf("expected ErrCheckpointInFuture, got %v", err)
	}
}

func TestMovementWithoutAShippingMethodIsRefused(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Owner Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	created, err := batches.Create(context.Background(), owner, declaration("batch-key-method"))
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	entry := checkpointEntry(domain.DepartedOrigin, "checkpoint-key-method")
	entry.ShippingMethod = ""

	if _, err := batches.LogCheckpoint(context.Background(), owner, created.Batch.ID, entry); !errors.Is(err, service.ErrShippingMethodRequired) {
		t.Fatalf("expected ErrShippingMethodRequired, got %v", err)
	}
}

func TestCorrectionSupersedesTheOriginalAndCannotRepeat(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Owner Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	created, err := batches.Create(context.Background(), owner, declaration("batch-key-correct"))
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	original, err := batches.LogCheckpoint(
		context.Background(), owner, created.Batch.ID,
		checkpointEntry(domain.ProductionComplete, "checkpoint-key-original"),
	)
	if err != nil {
		t.Fatalf("log original: %v", err)
	}

	correction := checkpointEntry(domain.ProductionComplete, "checkpoint-key-correction")
	correction.CorrectsID = &original.Checkpoint.ID
	correction.CorrectionReason = "Reported the wrong fab line"

	if _, err := batches.LogCheckpoint(context.Background(), owner, created.Batch.ID, correction); err != nil {
		t.Fatalf("file correction: %v", err)
	}

	second := checkpointEntry(domain.ProductionComplete, "checkpoint-key-correction-2")
	second.CorrectsID = &original.Checkpoint.ID
	second.CorrectionReason = "Changed my mind again"

	if _, err := batches.LogCheckpoint(context.Background(), owner, created.Batch.ID, second); !errors.Is(err, service.ErrAlreadyCorrected) {
		t.Fatalf("expected ErrAlreadyCorrected, got %v", err)
	}

	view, err := batches.Get(context.Background(), owner, created.Batch.ID)
	if err != nil {
		t.Fatalf("get batch: %v", err)
	}
	if view.Batch.CheckpointCount != 1 {
		t.Fatalf("a superseded checkpoint must not be counted twice, got %d", view.Batch.CheckpointCount)
	}
}

func TestCorrectionNeedsAReason(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Owner Org")
	batches := batchService(handle, stubFacilities{name: "Fab"})

	created, err := batches.Create(context.Background(), owner, declaration("batch-key-reason"))
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}

	original, err := batches.LogCheckpoint(
		context.Background(), owner, created.Batch.ID,
		checkpointEntry(domain.ProductionComplete, "checkpoint-key-reason-original"),
	)
	if err != nil {
		t.Fatalf("log original: %v", err)
	}

	correction := checkpointEntry(domain.ProductionComplete, "checkpoint-key-reason-correction")
	correction.CorrectsID = &original.Checkpoint.ID

	if _, err := batches.LogCheckpoint(context.Background(), owner, created.Batch.ID, correction); !errors.Is(err, service.ErrCorrectionReasonRequired) {
		t.Fatalf("expected ErrCorrectionReasonRequired, got %v", err)
	}
}

func TestCreditBuyerAndReadOnlyOrganizationsCannotProduce(t *testing.T) {
	handle := store(t)
	batches := batchService(handle, stubFacilities{name: "Fab"})

	buyer := actor(t, handle, "Buyer Org")
	buyer.OrganizationType = "credit_buyer"
	if _, err := batches.Create(context.Background(), buyer, declaration("batch-key-buyer")); !errors.Is(err, service.ErrPlanForbidsBatches) {
		t.Fatalf("expected ErrPlanForbidsBatches, got %v", err)
	}

	frozen := actor(t, handle, "Frozen Org")
	frozen.OrganizationState = "read_only"
	if _, err := batches.Create(context.Background(), frozen, declaration("batch-key-frozen")); !errors.Is(err, service.ErrOrganizationReadOnly) {
		t.Fatalf("expected ErrOrganizationReadOnly, got %v", err)
	}
}

func TestUnknownFacilityIsRefused(t *testing.T) {
	handle := store(t)
	owner := actor(t, handle, "Owner Org")
	batches := batchService(handle, stubFacilities{missing: true})

	if _, err := batches.Create(context.Background(), owner, declaration("batch-key-nofacility")); !errors.Is(err, service.ErrFacilityUnknown) {
		t.Fatalf("expected ErrFacilityUnknown, got %v", err)
	}
}
