package rpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	evidencev1 "github.com/carboncircuit/backend/gen/carboncircuit/evidence/v1"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
	"github.com/carboncircuit/backend/services/evidence-service/internal/service"
)

type DocumentManager interface {
	Upload(ctx context.Context, actor service.Actor, submission service.Submission) (service.DocumentView, error)
	Get(ctx context.Context, actor service.Actor, documentID uuid.UUID) (domain.Document, error)
	List(ctx context.Context, actor service.Actor, purpose domain.Purpose, after string, limit int) (service.Page, error)
	DownloadLink(ctx context.Context, actor service.Actor, documentID uuid.UUID) (service.Link, domain.Document, error)
	Resolve(ctx context.Context, actor service.Actor, documentIDs []uuid.UUID, purpose domain.Purpose) ([]service.Resolution, error)
}

type Reachable interface {
	Reachable(ctx context.Context) bool
}

type EvidenceServer struct {
	evidencev1.UnimplementedEvidenceServiceServer

	database  *gorm.DB
	documents DocumentManager
	objects   Reachable
	scanner   string
	sizeLimit int64
	logger    *slog.Logger
	revision  string
}

func NewEvidenceServer(
	database *gorm.DB,
	documents DocumentManager,
	objects Reachable,
	scanner string,
	sizeLimit int64,
	logger *slog.Logger,
	revision string,
) *EvidenceServer {
	return &EvidenceServer{
		database:  database,
		documents: documents,
		objects:   objects,
		scanner:   scanner,
		sizeLimit: sizeLimit,
		logger:    logger,
		revision:  revision,
	}
}

func (s *EvidenceServer) Ping(
	ctx context.Context,
	_ *evidencev1.PingRequest,
) (*evidencev1.PingResponse, error) {
	return &evidencev1.PingResponse{
		Service:           "evidence-service",
		Revision:          s.revision,
		DatabaseReachable: s.databaseReachable(ctx),
		StorageReachable:  s.objects.Reachable(ctx),
		Scanner:           s.scanner,
	}, nil
}

func (s *EvidenceServer) databaseReachable(ctx context.Context) bool {
	pool, err := s.database.DB()
	if err != nil {
		return false
	}
	return pool.PingContext(ctx) == nil
}

func (s *EvidenceServer) UploadDocument(
	stream evidencev1.EvidenceService_UploadDocumentServer,
) error {
	actor, err := s.actor(stream.Context())
	if err != nil {
		return err
	}

	submission, err := s.receive(stream)
	if err != nil {
		return err
	}

	view, err := s.documents.Upload(stream.Context(), actor, submission)
	if err != nil {
		return translate(err)
	}

	return stream.SendAndClose(&evidencev1.UploadDocumentResponse{
		Document:        documentMessage(view.Document),
		AlreadyUploaded: view.Replayed,
	})
}

func (s *EvidenceServer) receive(
	stream evidencev1.EvidenceService_UploadDocumentServer,
) (service.Submission, error) {
	var (
		submission service.Submission
		content    []byte
		described  bool
	)

	for {
		part, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return service.Submission{}, status.Error(codes.Canceled, "the upload stream ended early")
		}

		switch payload := part.GetPart().(type) {
		case *evidencev1.UploadDocumentRequest_Metadata:
			if described {
				return service.Submission{}, status.Error(codes.InvalidArgument,
					"upload metadata may only be sent once, as the first message")
			}
			described = true
			submission, err = submissionFrom(payload.Metadata)
			if err != nil {
				return service.Submission{}, err
			}
		case *evidencev1.UploadDocumentRequest_Chunk:
			if !described {
				return service.Submission{}, status.Error(codes.InvalidArgument,
					"the first message of an upload must be its metadata")
			}
			if int64(len(content))+int64(len(payload.Chunk)) > s.sizeLimit {
				return service.Submission{}, status.Errorf(codes.InvalidArgument,
					"evidence may not exceed %d bytes", s.sizeLimit)
			}
			content = append(content, payload.Chunk...)
		}
	}

	if !described {
		return service.Submission{}, status.Error(codes.InvalidArgument, "the upload carried no metadata")
	}
	if len(content) == 0 {
		return service.Submission{}, status.Error(codes.InvalidArgument, "the upload carried no content")
	}

	submission.Content = content

	return submission, nil
}

func submissionFrom(metadata *evidencev1.UploadMetadata) (service.Submission, error) {
	if metadata.GetIdempotencyKey() == "" {
		return service.Submission{}, status.Error(codes.InvalidArgument, "an idempotency key is required")
	}

	purpose, known := purposeFrom(metadata.GetPurpose())
	if !known {
		return service.Submission{}, status.Error(codes.InvalidArgument, "purpose must be a known evidence purpose")
	}

	return service.Submission{
		Purpose:           purpose,
		FileName:          metadata.GetFileName(),
		DeclaredMediaType: metadata.GetDeclaredMediaType(),
		IdempotencyKey:    metadata.GetIdempotencyKey(),
	}, nil
}

func (s *EvidenceServer) GetDocument(
	ctx context.Context,
	request *evidencev1.GetDocumentRequest,
) (*evidencev1.GetDocumentResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(request.GetDocumentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "document id is not a valid identifier")
	}

	document, err := s.documents.Get(ctx, actor, documentID)
	if err != nil {
		return nil, translate(err)
	}

	return &evidencev1.GetDocumentResponse{Document: documentMessage(document)}, nil
}

func (s *EvidenceServer) ListDocuments(
	ctx context.Context,
	request *evidencev1.ListDocumentsRequest,
) (*evidencev1.ListDocumentsResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	purpose, _ := purposeFrom(request.GetPurpose())

	page, err := s.documents.List(ctx, actor, purpose, request.GetAfter(), int(request.GetLimit()))
	if err != nil {
		return nil, translate(err)
	}

	documents := make([]*evidencev1.Document, 0, len(page.Documents))
	for _, document := range page.Documents {
		documents = append(documents, documentMessage(document))
	}

	return &evidencev1.ListDocumentsResponse{
		Documents:  documents,
		NextCursor: page.NextCursor,
	}, nil
}

func (s *EvidenceServer) CreateDownloadLink(
	ctx context.Context,
	request *evidencev1.CreateDownloadLinkRequest,
) (*evidencev1.CreateDownloadLinkResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(request.GetDocumentId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "document id is not a valid identifier")
	}

	link, document, err := s.documents.DownloadLink(ctx, actor, documentID)
	if err != nil {
		return nil, translate(err)
	}

	return &evidencev1.CreateDownloadLinkResponse{
		Url:       link.URL,
		ExpiresAt: link.ExpiresAt.UTC().Format(time.RFC3339),
		FileName:  document.FileName,
		MediaType: document.DetectedMediaType,
	}, nil
}

func (s *EvidenceServer) ResolveDocuments(
	ctx context.Context,
	request *evidencev1.ResolveDocumentsRequest,
) (*evidencev1.ResolveDocumentsResponse, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return nil, err
	}

	documentIDs := make([]uuid.UUID, 0, len(request.GetDocumentIds()))
	for _, candidate := range request.GetDocumentIds() {
		parsed, err := uuid.Parse(candidate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument,
				"document id %q is not a valid identifier", candidate)
		}
		documentIDs = append(documentIDs, parsed)
	}

	purpose, _ := purposeFrom(request.GetPurpose())

	resolutions, err := s.documents.Resolve(ctx, actor, documentIDs, purpose)
	if err != nil {
		return nil, translate(err)
	}

	resolved := make([]*evidencev1.ResolvedDocument, 0, len(resolutions))
	for _, resolution := range resolutions {
		resolved = append(resolved, resolvedMessage(resolution))
	}

	return &evidencev1.ResolveDocumentsResponse{Documents: resolved}, nil
}

func (s *EvidenceServer) actor(ctx context.Context) (service.Actor, error) {
	verified, present := grpcx.CallerFrom(ctx)
	if !present || !verified.HasOrganization() {
		return service.Actor{}, status.Error(codes.Unauthenticated, "a verified organization is required")
	}
	return actorFrom(verified)
}

func actorFrom(verified servicetoken.Caller) (service.Actor, error) {
	organizationID, err := uuid.Parse(verified.OrganizationID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable organization")
	}

	userID, err := uuid.Parse(verified.UserID)
	if err != nil {
		return service.Actor{}, status.Error(codes.Unauthenticated, "service token carries an unusable user")
	}

	return service.Actor{
		OrganizationID:    organizationID,
		UserID:            userID,
		OrganizationState: verified.OrganizationState,
	}, nil
}

func translate(err error) error {
	var refusal *service.Refusal
	if errors.As(err, &refusal) {
		return refused(refusal)
	}

	switch {
	case errors.Is(err, service.ErrDocumentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrOrganizationReadOnly):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, service.ErrNotDownloadable):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrRequestInProgress):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, service.ErrIdempotencyConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrPurposeUnknown),
		errors.Is(err, service.ErrFileNameRequired):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func refused(refusal *service.Refusal) error {
	reported := status.New(codes.InvalidArgument, refusal.Reason)

	detailed, err := reported.WithDetails(&errdetails.ErrorInfo{
		Reason:   strings.ToUpper(string(refusal.Verdict)),
		Domain:   events.EvidenceRefusalDomain,
		Metadata: map[string]string{"document_id": refusal.Document.ID.String()},
	})
	if err != nil {
		return reported.Err()
	}

	return detailed.Err()
}
