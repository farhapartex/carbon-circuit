package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
	"github.com/carboncircuit/backend/services/evidence-service/internal/inspect"
	"github.com/carboncircuit/backend/services/evidence-service/internal/repository"
	"github.com/carboncircuit/backend/services/evidence-service/internal/scan"
)

const (
	uploadEndpoint    = "POST /v1/evidence"
	documentAggregate = "evidence"
	defaultPageSize   = 25
	maximumPageSize   = 100
)

var (
	ErrRequestInProgress    = errors.New("an identical request is already being processed")
	ErrIdempotencyConflict  = errors.New("idempotency key was used with a different request")
	ErrDocumentNotFound     = errors.New("document not found")
	ErrOrganizationReadOnly = errors.New("this organization cannot upload evidence right now")
	ErrPurposeUnknown       = errors.New("upload purpose is not a known evidence purpose")
	ErrFileNameRequired     = errors.New("file name is required")
	ErrNotDownloadable      = errors.New("a document is only downloadable once it has passed scanning")
)

type Refusal struct {
	Verdict  domain.ScanVerdict
	Reason   string
	Document domain.Document
}

func (r *Refusal) Error() string {
	return fmt.Sprintf("evidence refused (%s): %s", r.Verdict, r.Reason)
}

var purposes = map[domain.Purpose]struct{}{
	domain.ClaimEvidence:      {},
	domain.BatchCertification: {},
}

type Actor struct {
	OrganizationID    uuid.UUID
	UserID            uuid.UUID
	OrganizationState string
}

func (a Actor) mayUpload() error {
	if a.OrganizationState == "read_only" || a.OrganizationState == "suspended" {
		return ErrOrganizationReadOnly
	}
	return nil
}

type Submission struct {
	Purpose           domain.Purpose
	FileName          string
	DeclaredMediaType string
	Content           []byte
	IdempotencyKey    string
}

func (s Submission) validate() error {
	if _, known := purposes[s.Purpose]; !known {
		return ErrPurposeUnknown
	}
	if strings.TrimSpace(s.FileName) == "" {
		return ErrFileNameRequired
	}
	return nil
}

type ObjectStore interface {
	Put(ctx context.Context, key string, content []byte, mediaType string) error
	Remove(ctx context.Context, key string) error
	SignedLink(ctx context.Context, key, fileName, mediaType string) (string, time.Time, error)
	Reachable(ctx context.Context) bool
}

type Link struct {
	URL       string
	ExpiresAt time.Time
}

type DocumentView struct {
	Document domain.Document
	Replayed bool
}

type DocumentService struct {
	database  *gorm.DB
	documents repository.DocumentStore
	objects   ObjectStore
	scanner   scan.Scanner
	limits    inspect.Limits
	logger    *slog.Logger
}

func NewDocumentService(
	handle *gorm.DB,
	documents repository.DocumentStore,
	objects ObjectStore,
	scanner scan.Scanner,
	limits inspect.Limits,
	logger *slog.Logger,
) *DocumentService {
	return &DocumentService{
		database:  handle,
		documents: documents,
		objects:   objects,
		scanner:   scanner,
		limits:    limits,
		logger:    logger,
	}
}

func (s *DocumentService) Upload(
	ctx context.Context,
	actor Actor,
	submission Submission,
) (DocumentView, error) {
	if err := actor.mayUpload(); err != nil {
		return DocumentView{}, err
	}
	if err := submission.validate(); err != nil {
		return DocumentView{}, err
	}

	documentID, err := uuid.NewV7()
	if err != nil {
		return DocumentView{}, fmt.Errorf("generate document id: %w", err)
	}

	examined, verdict, reason := s.examine(ctx, submission)
	if verdict != domain.Accepted {
		return DocumentView{}, s.refuse(ctx, actor, documentID, submission, verdict, reason)
	}

	storageKey := objectKey(actor.OrganizationID, documentID, examined.MediaType)
	storedHash := digestOf(examined.Content)

	if err := s.objects.Put(ctx, storageKey, examined.Content, examined.MediaType); err != nil {
		return DocumentView{}, err
	}

	document := domain.Document{
		OrganizationID:        actor.OrganizationID,
		UploadedByUserID:      actor.UserID,
		Purpose:               submission.Purpose,
		FileName:              path.Base(strings.TrimSpace(submission.FileName)),
		DeclaredMediaType:     submission.DeclaredMediaType,
		DetectedMediaType:     examined.MediaType,
		ByteSize:              int64(len(examined.Content)),
		PageCount:             examined.PageCount,
		ContentHash:           digestOf(submission.Content),
		StoredHash:            &storedHash,
		StorageKey:            &storageKey,
		ScanStatus:            domain.ScanClean,
		ScanVerdict:           domain.Accepted,
		ScannedBy:             s.scanner.Name(),
		ScannedAt:             time.Now().UTC(),
		ActiveContentStripped: examined.ActiveContentStripped,
	}
	document.ID = documentID

	view, err := s.commit(ctx, actor, document, submission)
	if err != nil {
		s.discard(ctx, storageKey)
		return DocumentView{}, err
	}

	return view, nil
}

func (s *DocumentService) examine(
	ctx context.Context,
	submission Submission,
) (inspect.Result, domain.ScanVerdict, string) {
	examined, err := inspect.Inspect(submission.DeclaredMediaType, submission.Content, s.limits)
	if err != nil {
		return inspect.Result{}, verdictFor(err), err.Error()
	}

	outcome, err := s.scanner.Scan(ctx, examined.Content)
	if err != nil {
		return inspect.Result{}, domain.RejectedMalware,
			fmt.Sprintf("the malware scanner could not reach a verdict: %v", err)
	}
	if !outcome.Clean {
		return inspect.Result{}, domain.RejectedMalware,
			fmt.Sprintf("malware signature %s", outcome.Signature)
	}

	return examined, domain.Accepted, ""
}

func verdictFor(err error) domain.ScanVerdict {
	switch {
	case errors.Is(err, inspect.ErrMediaTypeNotAccepted):
		return domain.RejectedMediaType
	case errors.Is(err, inspect.ErrContentMismatch):
		return domain.RejectedContentMismatch
	case errors.Is(err, inspect.ErrTooLarge):
		return domain.RejectedTooLarge
	case errors.Is(err, inspect.ErrPageCountExceeded):
		return domain.RejectedPageCount
	case errors.Is(err, inspect.ErrActiveContent):
		return domain.RejectedActiveContent
	default:
		return domain.RejectedUnreadable
	}
}

func (s *DocumentService) refuse(
	ctx context.Context,
	actor Actor,
	documentID uuid.UUID,
	submission Submission,
	verdict domain.ScanVerdict,
	reason string,
) error {
	document := domain.Document{
		OrganizationID:    actor.OrganizationID,
		UploadedByUserID:  actor.UserID,
		Purpose:           submission.Purpose,
		FileName:          path.Base(strings.TrimSpace(submission.FileName)),
		DeclaredMediaType: submission.DeclaredMediaType,
		DetectedMediaType: submission.DeclaredMediaType,
		ByteSize:          int64(len(submission.Content)),
		ContentHash:       digestOf(submission.Content),
		ScanStatus:        domain.ScanFailed,
		ScanVerdict:       verdict,
		ScannedBy:         s.scanner.Name(),
		ScannedAt:         time.Now().UTC(),
	}
	document.ID = documentID

	err := database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		if err := s.documents.Insert(tx, &document); err != nil {
			return err
		}
		return publish(tx, document)
	})
	if err != nil {
		s.logger.Error("could not record a refused evidence upload",
			slog.String("document_id", documentID.String()),
			slog.String("verdict", string(verdict)),
			slog.String("error", err.Error()))
	}

	return &Refusal{Verdict: verdict, Reason: reason, Document: document}
}

func (s *DocumentService) commit(
	ctx context.Context,
	actor Actor,
	document domain.Document,
	submission Submission,
) (DocumentView, error) {
	var view DocumentView

	err := database.WithinTenant(ctx, s.database, tenancy(actor), func(tx database.Tx) error {
		reservation, err := idempotency.Reserve(tx, idempotency.Request{
			Scope:    idempotency.ForOrganization(actor.OrganizationID),
			Endpoint: uploadEndpoint,
			Key:      submission.IdempotencyKey,
			Body:     idempotentBody(submission, document.ContentHash),
		})
		switch {
		case errors.Is(err, idempotency.ErrInProgress):
			return ErrRequestInProgress
		case errors.Is(err, idempotency.ErrKeyReused):
			return ErrIdempotencyConflict
		case err != nil:
			return err
		}

		if reservation.IsReplay() {
			var replayed domain.Document
			if err := json.Unmarshal(reservation.Replay.Body, &replayed); err != nil {
				return fmt.Errorf("decode replayed document: %w", err)
			}
			view = DocumentView{Document: replayed, Replayed: true}
			return nil
		}

		if err := s.documents.Insert(tx, &document); err != nil {
			return err
		}
		if err := publish(tx, document); err != nil {
			return err
		}

		body, err := json.Marshal(document)
		if err != nil {
			return fmt.Errorf("encode idempotent response: %w", err)
		}

		if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
			Status:     201,
			Body:       body,
			ResourceID: &document.ID,
		}); err != nil {
			return err
		}

		view = DocumentView{Document: document}
		return nil
	})
	if err != nil {
		return DocumentView{}, err
	}

	return view, nil
}

func (s *DocumentService) discard(ctx context.Context, storageKey string) {
	if err := s.objects.Remove(ctx, storageKey); err != nil {
		s.logger.Error("stored object outlived its failed transaction",
			slog.String("storage_key", storageKey),
			slog.String("error", err.Error()))
	}
}

func publish(tx database.Tx, document domain.Document) error {
	pages := 0
	if document.PageCount != nil {
		pages = *document.PageCount
	}

	_, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: documentAggregate,
		AggregateID:   document.ID,
		EventType:     events.TopicEvidenceScanned,
		Payload: events.EvidenceScanned{
			EvidenceID:       document.ID.String(),
			OrganizationID:   document.OrganizationID.String(),
			Purpose:          string(document.Purpose),
			ContentHash:      document.ContentHash,
			MediaType:        document.DetectedMediaType,
			ByteSize:         document.ByteSize,
			PageCount:        pages,
			ScanStatus:       string(document.ScanStatus),
			ScanVerdict:      string(document.ScanVerdict),
			ScannedBy:        document.ScannedBy,
			ScannedAt:        document.ScannedAt.UTC().Format(time.RFC3339),
			UploadedByUserID: document.UploadedByUserID.String(),
		},
	})

	return err
}

func tenancy(actor Actor) database.TenantContext {
	return database.TenantContext{
		UserID:         actor.UserID.String(),
		OrganizationID: actor.OrganizationID.String(),
	}
}

func objectKey(organizationID, documentID uuid.UUID, mediaType string) string {
	return fmt.Sprintf("%s/%s%s", organizationID, documentID, extensionFor(mediaType))
}

func extensionFor(mediaType string) string {
	switch mediaType {
	case inspect.PDF:
		return ".pdf"
	case inspect.PNG:
		return ".png"
	case inspect.JPEG:
		return ".jpg"
	case inspect.CSV:
		return ".csv"
	case inspect.XLSX:
		return ".xlsx"
	default:
		return ""
	}
}

func idempotentBody(submission Submission, contentHash string) []byte {
	return []byte(strings.Join([]string{
		string(submission.Purpose),
		path.Base(strings.TrimSpace(submission.FileName)),
		contentHash,
	}, "\n"))
}

func digestOf(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}
