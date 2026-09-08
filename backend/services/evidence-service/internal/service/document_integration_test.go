package service_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
	"github.com/carboncircuit/backend/services/evidence-service/internal/inspect"
	"github.com/carboncircuit/backend/services/evidence-service/internal/service"
)

func TestAcceptedUploadIsStoredHashedAndPublished(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	statement := statementPDF(t, "")

	view, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "q1-electricity.pdf", inspect.PDF, statement, "upload-1"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if view.Document.ScanStatus != domain.ScanClean {
		t.Fatalf("expected a clean document, got %s", view.Document.ScanStatus)
	}
	if view.Document.StorageKey == nil {
		t.Fatal("a clean document must carry a storage key")
	}
	if len(view.Document.ContentHash) != 64 {
		t.Fatalf("expected a sha256 hex digest, got %q", view.Document.ContentHash)
	}
	if view.Document.PageCount == nil || *view.Document.PageCount != 1 {
		t.Fatalf("expected one page, got %v", view.Document.PageCount)
	}
	if objects.count() != 1 {
		t.Fatalf("expected exactly one stored object, got %d", objects.count())
	}

	if view.Document.ContentHash != sha256Hex(statement) {
		t.Fatal("content_hash must describe the bytes as submitted, since that is the hash recorded publicly")
	}

	stored, present := objects.content(*view.Document.StorageKey)
	if !present {
		t.Fatal("the storage key does not point at a stored object")
	}
	if view.Document.StoredHash == nil || *view.Document.StoredHash != sha256Hex(stored) {
		t.Fatal("stored_hash must describe the bytes actually written to storage")
	}

	if published := outboxCount(t, handle, uploader, view.Document.ID); published != 1 {
		t.Fatalf("expected one evidence.scanned event, got %d", published)
	}
}

func TestIdenticalSubmissionsShareAContentHashDespiteSanitization(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	statement := statementPDF(t, "")

	first, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "a.pdf", inspect.PDF, statement, "dedupe-a"))
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}

	second, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "b.pdf", inspect.PDF, statement, "dedupe-b"))
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}

	if first.Document.ContentHash != second.Document.ContentHash {
		t.Fatal("the same file submitted twice must share a content hash, or the duplicate-evidence rule cannot fire")
	}
	if *first.Document.StoredHash == *second.Document.StoredHash {
		t.Skip("sanitization happened to be byte-identical, so this run proves nothing about stored_hash")
	}
}

func TestSanitizedBytesAreWhatGetsStored(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	hostile := statementPDF(t, " /OpenAction << /S /JavaScript /JS (app.alert\\('owned'\\);) >>")

	view, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "hostile.pdf", inspect.PDF, hostile, "upload-hostile"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if !view.Document.ActiveContentStripped {
		t.Fatal("expected the upload to record that active content was stripped")
	}

	stored, present := objects.content(*view.Document.StorageKey)
	if !present {
		t.Fatal("the storage key does not point at a stored object")
	}
	if bytes.Contains(stored, []byte("app.alert")) {
		t.Fatal("the bytes in storage still carry the script body")
	}
	if bytes.Equal(stored, hostile) {
		t.Fatal("the submitted bytes were stored verbatim, so nothing was stripped")
	}
	if *view.Document.StoredHash != sha256Hex(stored) {
		t.Fatal("stored_hash does not describe the sanitized bytes in storage")
	}
	if view.Document.ContentHash != sha256Hex(hostile) {
		t.Fatal("content_hash must still describe what the uploader submitted")
	}
}

func TestRefusedUploadStoresNothingButIsRecorded(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	_, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "book.xlsm", inspect.XLSX, macroWorkbook(t), "upload-macro"))

	var refusal *service.Refusal
	if !errors.As(err, &refusal) {
		t.Fatalf("expected a Refusal, got %v", err)
	}
	if refusal.Verdict != domain.RejectedActiveContent {
		t.Fatalf("expected rejected_active_content, got %s", refusal.Verdict)
	}
	if objects.count() != 0 {
		t.Fatal("a refused upload must never reach object storage")
	}

	recorded, getErr := documents.Get(context.Background(), uploader, refusal.Document.ID)
	if getErr != nil {
		t.Fatalf("a refused upload must still be auditable: %v", getErr)
	}
	if recorded.ScanStatus != domain.ScanFailed {
		t.Fatalf("expected scan_status failed, got %s", recorded.ScanStatus)
	}
	if recorded.StorageKey != nil {
		t.Fatal("a refused upload must not carry a storage key")
	}
	if recorded.Usable() {
		t.Fatal("a refused upload must never report itself usable")
	}
}

func TestMalwareVerdictIsRefusedAndRecorded(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, refusingScanner{signature: "Eicar-Test-Signature"})
	uploader := actor(t, handle)

	_, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "invoice.pdf", inspect.PDF, statementPDF(t, ""), "upload-malware"))

	var refusal *service.Refusal
	if !errors.As(err, &refusal) {
		t.Fatalf("expected a Refusal, got %v", err)
	}
	if refusal.Verdict != domain.RejectedMalware {
		t.Fatalf("expected rejected_malware, got %s", refusal.Verdict)
	}
	if objects.count() != 0 {
		t.Fatal("a document with a malware signature must never be stored")
	}
	if refusal.Document.ScannedBy != "probe-scanner" {
		t.Fatalf("the row must record which scanner ran, got %q", refusal.Document.ScannedBy)
	}
}

func TestUnreachableScannerRefusesRatherThanTrusts(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects,
		refusingScanner{failure: errors.New("connection refused")})
	uploader := actor(t, handle)

	_, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "invoice.pdf", inspect.PDF, statementPDF(t, ""), "upload-scanner-down"))

	var refusal *service.Refusal
	if !errors.As(err, &refusal) {
		t.Fatalf("a scanner that cannot reach a verdict must refuse the upload, got %v", err)
	}
	if objects.count() != 0 {
		t.Fatal("an unscanned document must never be stored")
	}
}

func TestOversizeUploadIsRefused(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	oversize := make([]byte, domain.MaximumByteSize+1)
	copy(oversize, statementPDF(t, ""))

	_, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "huge.pdf", inspect.PDF, oversize, "upload-huge"))

	var refusal *service.Refusal
	if !errors.As(err, &refusal) {
		t.Fatalf("expected a Refusal, got %v", err)
	}
	if refusal.Verdict != domain.RejectedTooLarge {
		t.Fatalf("expected rejected_too_large, got %s", refusal.Verdict)
	}
}

func TestReplayReturnsTheOriginalDocument(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	statement := statementPDF(t, "")
	request := submission(domain.ClaimEvidence, "q1.pdf", inspect.PDF, statement, "upload-replay")

	first, err := documents.Upload(context.Background(), uploader, request)
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}

	second, err := documents.Upload(context.Background(), uploader, request)
	if err != nil {
		t.Fatalf("replayed upload: %v", err)
	}

	if !second.Replayed {
		t.Fatal("expected the second upload to be recognised as a replay")
	}
	if second.Document.ID != first.Document.ID {
		t.Fatalf("replay produced a new document %s, expected %s",
			second.Document.ID, first.Document.ID)
	}
	if documentCount(t, handle, uploader) != 1 {
		t.Fatal("a replayed upload must not produce a second row")
	}
}

func TestAnotherOrganizationCannotSeeOrDownloadTheDocument(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())

	owner := actor(t, handle)
	stranger := actor(t, handle)

	view, err := documents.Upload(context.Background(), owner,
		submission(domain.ClaimEvidence, "audit.pdf", inspect.PDF, statementPDF(t, ""), "upload-owned"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if _, err := documents.Get(context.Background(), stranger, view.Document.ID); !errors.Is(err, service.ErrDocumentNotFound) {
		t.Fatalf("expected ErrDocumentNotFound for another organization, got %v", err)
	}

	if _, _, err := documents.DownloadLink(context.Background(), stranger, view.Document.ID); !errors.Is(err, service.ErrDocumentNotFound) {
		t.Fatalf("expected a stranger to be refused a download link, got %v", err)
	}

	page, err := documents.List(context.Background(), stranger, "", "", 25)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Documents) != 0 {
		t.Fatalf("a stranger listed %d documents belonging to another organization", len(page.Documents))
	}
}

func TestDownloadLinkIsIssuedOnlyForACleanDocument(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	uploader := actor(t, handle)

	clean := documentService(handle, objects, cleanScanner())
	view, err := clean.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "audit.pdf", inspect.PDF, statementPDF(t, ""), "upload-clean"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	link, document, err := clean.DownloadLink(context.Background(), uploader, view.Document.ID)
	if err != nil {
		t.Fatalf("download link: %v", err)
	}
	if link.URL == "" || link.ExpiresAt.IsZero() {
		t.Fatal("expected a signed link with an expiry")
	}
	if document.FileName != "audit.pdf" {
		t.Fatalf("expected the original file name, got %q", document.FileName)
	}

	refused := documentService(handle, objects, refusingScanner{signature: "Probe"})
	_, uploadErr := refused.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "bad.pdf", inspect.PDF, statementPDF(t, ""), "upload-dirty"))

	var refusal *service.Refusal
	if !errors.As(uploadErr, &refusal) {
		t.Fatalf("expected a Refusal, got %v", uploadErr)
	}

	if _, _, err := clean.DownloadLink(context.Background(), uploader, refusal.Document.ID); !errors.Is(err, service.ErrNotDownloadable) {
		t.Fatalf("expected ErrNotDownloadable for a failed scan, got %v", err)
	}
}

func TestResolveRefusesEverythingThatIsNotACleanOwnedDocument(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())

	owner := actor(t, handle)
	stranger := actor(t, handle)

	clean, err := documents.Upload(context.Background(), owner,
		submission(domain.ClaimEvidence, "audit.pdf", inspect.PDF, statementPDF(t, ""), "resolve-clean"))
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	certification, err := documents.Upload(context.Background(), owner,
		submission(domain.BatchCertification, "cert.pdf", inspect.PDF, statementPDF(t, " /Lang (en)"), "resolve-cert"))
	if err != nil {
		t.Fatalf("upload certification: %v", err)
	}

	refusedService := documentService(handle, objects, refusingScanner{signature: "Probe"})
	_, uploadErr := refusedService.Upload(context.Background(), owner,
		submission(domain.ClaimEvidence, "bad.pdf", inspect.PDF, statementPDF(t, " /Lang (fr)"), "resolve-dirty"))

	var refusal *service.Refusal
	if !errors.As(uploadErr, &refusal) {
		t.Fatalf("expected a Refusal, got %v", uploadErr)
	}

	unknown := uuid.New()

	resolutions, err := documents.Resolve(context.Background(), owner,
		[]uuid.UUID{clean.Document.ID, certification.Document.ID, refusal.Document.ID, unknown},
		domain.ClaimEvidence)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(resolutions) != 4 {
		t.Fatalf("expected one resolution per requested id, got %d", len(resolutions))
	}

	if !resolutions[0].Usable {
		t.Fatalf("a clean owned claim document must be usable, got %q", resolutions[0].Refusal)
	}
	if resolutions[1].Usable {
		t.Fatal("a batch certification must not be usable as claim evidence")
	}
	if resolutions[2].Usable {
		t.Fatal("a document that failed scanning must never be usable")
	}
	if resolutions[3].Usable {
		t.Fatal("an unknown document id must never be usable")
	}

	strangerView, err := documents.Resolve(context.Background(), stranger,
		[]uuid.UUID{clean.Document.ID}, domain.ClaimEvidence)
	if err != nil {
		t.Fatalf("resolve as stranger: %v", err)
	}
	if strangerView[0].Usable {
		t.Fatal("another organization's document must never resolve as usable")
	}
}

func TestReadOnlyOrganizationCannotUpload(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())

	restricted := actor(t, handle)
	restricted.OrganizationState = "read_only"

	_, err := documents.Upload(context.Background(), restricted,
		submission(domain.ClaimEvidence, "audit.pdf", inspect.PDF, statementPDF(t, ""), "upload-restricted"))
	if !errors.Is(err, service.ErrOrganizationReadOnly) {
		t.Fatalf("expected ErrOrganizationReadOnly, got %v", err)
	}
	if objects.count() != 0 {
		t.Fatal("a restricted organization must not reach object storage")
	}
}

func TestUnknownPurposeIsRefusedBeforeAnyWork(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	_, err := documents.Upload(context.Background(), uploader,
		submission("invoice", "audit.pdf", inspect.PDF, statementPDF(t, ""), "upload-bad-purpose"))
	if !errors.Is(err, service.ErrPurposeUnknown) {
		t.Fatalf("expected ErrPurposeUnknown, got %v", err)
	}
	if objects.count() != 0 {
		t.Fatal("an unknown purpose must not reach object storage")
	}
}

func TestFailedTransactionLeavesNoOrphanedObject(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	accepted, err := documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "first.pdf", inspect.PDF, statementPDF(t, ""), "shared-key"))
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}

	_, err = documents.Upload(context.Background(), uploader,
		submission(domain.ClaimEvidence, "second.pdf", inspect.PDF, statementPDF(t, " /Lang (en)"), "shared-key"))
	if !errors.Is(err, service.ErrIdempotencyConflict) {
		t.Fatalf("expected ErrIdempotencyConflict, got %v", err)
	}

	if objects.count() != 1 {
		t.Fatalf("expected only the first document in storage, found %d objects", objects.count())
	}
	if _, present := objects.content(*accepted.Document.StorageKey); !present {
		t.Fatal("the compensating delete removed the wrong object")
	}
}

func TestListIsScopedAndOrdered(t *testing.T) {
	handle := store(t)
	objects := newMemoryObjects()
	documents := documentService(handle, objects, cleanScanner())
	uploader := actor(t, handle)

	for index, name := range []string{"one.pdf", "two.pdf", "three.pdf"} {
		_, err := documents.Upload(context.Background(), uploader,
			submission(domain.ClaimEvidence, name, inspect.PDF,
				statementPDF(t, " /Lang (l"+string(rune('a'+index))+")"), "list-"+name))
		if err != nil {
			t.Fatalf("upload %s: %v", name, err)
		}
	}

	page, err := documents.List(context.Background(), uploader, domain.ClaimEvidence, "", 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Documents) != 2 {
		t.Fatalf("expected a page of 2, got %d", len(page.Documents))
	}
	if page.NextCursor == "" {
		t.Fatal("expected a cursor for the remaining document")
	}
	if page.Documents[0].FileName != "three.pdf" {
		t.Fatalf("expected the newest document first, got %s", page.Documents[0].FileName)
	}

	rest, err := documents.List(context.Background(), uploader, domain.ClaimEvidence, page.NextCursor, 2)
	if err != nil {
		t.Fatalf("list rest: %v", err)
	}
	if len(rest.Documents) != 1 {
		t.Fatalf("expected the last document alone, got %d", len(rest.Documents))
	}
	if rest.Documents[0].FileName != "one.pdf" {
		t.Fatalf("expected one.pdf last, got %s", rest.Documents[0].FileName)
	}
}

func documentCount(t *testing.T, handle *gorm.DB, uploader service.Actor) int64 {
	t.Helper()

	var count int64

	err := database.WithinTenant(context.Background(), handle, database.TenantContext{
		UserID:         uploader.UserID.String(),
		OrganizationID: uploader.OrganizationID.String(),
	}, func(tx database.Tx) error {
		return tx.Session().
			Raw(`SELECT count(*) FROM evidence.documents WHERE organization_id = ?`,
				uploader.OrganizationID).
			Scan(&count).Error
	})
	if err != nil {
		t.Fatalf("count documents: %v", err)
	}

	return count
}

func outboxCount(t *testing.T, handle *gorm.DB, uploader service.Actor, documentID uuid.UUID) int64 {
	t.Helper()

	var count int64

	err := database.WithinTenant(context.Background(), handle, database.TenantContext{
		UserID:         uploader.UserID.String(),
		OrganizationID: uploader.OrganizationID.String(),
	}, func(tx database.Tx) error {
		return tx.Session().
			Raw(`SELECT count(*) FROM evidence.outbox_events WHERE aggregate_id = ? AND event_type = 'evidence.scanned'`,
				documentID).
			Scan(&count).Error
	})
	if err != nil {
		t.Fatalf("count outbox events: %v", err)
	}

	return count
}
