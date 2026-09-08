package service_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
	"github.com/carboncircuit/backend/services/evidence-service/internal/inspect"
	"github.com/carboncircuit/backend/services/evidence-service/internal/repository"
	"github.com/carboncircuit/backend/services/evidence-service/internal/scan"
	"github.com/carboncircuit/backend/services/evidence-service/internal/service"
)

var testLimits = inspect.Limits{MaximumBytes: 25 * 1024 * 1024, MaximumPages: 100}

type memoryObjects struct {
	mutex   sync.Mutex
	objects map[string][]byte
	putErr  error
}

func newMemoryObjects() *memoryObjects {
	return &memoryObjects{objects: map[string][]byte{}}
}

func (m *memoryObjects) Put(_ context.Context, key string, content []byte, _ string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.putErr != nil {
		return m.putErr
	}

	stored := make([]byte, len(content))
	copy(stored, content)
	m.objects[key] = stored

	return nil
}

func (m *memoryObjects) Remove(_ context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.objects, key)
	return nil
}

func (m *memoryObjects) SignedLink(_ context.Context, key, fileName, _ string) (service.Link, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, present := m.objects[key]; !present {
		return service.Link{}, fmt.Errorf("no object at %s", key)
	}

	return service.Link{
		URL:       "https://storage.invalid/" + key + "?filename=" + fileName,
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
	}, nil
}

func (m *memoryObjects) Reachable(context.Context) bool { return true }

func (m *memoryObjects) count() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return len(m.objects)
}

func (m *memoryObjects) content(key string) ([]byte, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	stored, present := m.objects[key]
	return stored, present
}

type refusingScanner struct {
	signature string
	failure   error
}

func (r refusingScanner) Name() string { return "probe-scanner" }

func (r refusingScanner) Scan(context.Context, []byte) (scan.Outcome, error) {
	if r.failure != nil {
		return scan.Outcome{}, r.failure
	}
	return scan.Outcome{Clean: false, Signature: r.signature}, nil
}

func store(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_DSN to run evidence integration tests")
	}

	opened, err := database.Open(context.Background(), database.Options{
		DSN:             dsn,
		Schema:          "evidence",
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

func documentService(
	handle *gorm.DB,
	objects service.ObjectStore,
	scanner scan.Scanner,
) *service.DocumentService {
	return service.NewDocumentService(
		handle,
		repository.NewDocumentRepository(),
		objects,
		scanner,
		testLimits,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func cleanScanner() scan.Scanner {
	scanner, err := scan.Open(scan.Disabled, "", slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		panic(err)
	}
	return scanner
}

func actor(t *testing.T, handle *gorm.DB) service.Actor {
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
				tx.Session().Exec(`DELETE FROM evidence.outbox_events WHERE aggregate_id IN (SELECT id FROM evidence.documents WHERE organization_id = ?)`, organizationID)
				tx.Session().Exec(`DELETE FROM evidence.idempotency_records WHERE organization_id = ?`, organizationID)
				return tx.Session().Exec(`DELETE FROM evidence.documents WHERE organization_id = ?`, organizationID).Error
			})
		if err != nil {
			t.Errorf("clean organization %s: %v", organizationID, err)
		}
	})

	return service.Actor{
		OrganizationID:    organizationID,
		UserID:            userID,
		OrganizationState: "active",
	}
}

type pdfBuilder struct {
	objects []string
}

func (b *pdfBuilder) bytes() []byte {
	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n")

	offsets := make([]int, len(b.objects))
	for index, body := range b.objects {
		offsets[index] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", index+1, body)
	}

	startXref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(b.objects)+1)
	out.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(b.objects)+1, startXref)

	return out.Bytes()
}

func statementPDF(t *testing.T, catalogExtras string) []byte {
	t.Helper()

	builder := &pdfBuilder{objects: []string{
		"<< /Type /Catalog /Pages 2 0 R" + catalogExtras + " >>",
		"<< /Type /Pages /Count 1 /Kids [3 0 R] >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>",
	}}

	return builder.bytes()
}

func macroWorkbook(t *testing.T) []byte {
	t.Helper()

	var out bytes.Buffer
	archive := zip.NewWriter(&out)

	for name, body := range map[string]string{
		"xl/workbook.xml":   `<?xml version="1.0"?><workbook/>`,
		"xl/vbaProject.bin": "\x00\x01macro payload",
	} {
		writer, err := archive.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		writer.Write([]byte(body))
	}

	if err := archive.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}

	return out.Bytes()
}

func submission(
	purpose domain.Purpose,
	fileName, mediaType string,
	content []byte,
	key string,
) service.Submission {
	return service.Submission{
		Purpose:           purpose,
		FileName:          fileName,
		DeclaredMediaType: mediaType,
		Content:           content,
		IdempotencyKey:    key,
	}
}

func sha256Hex(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}
