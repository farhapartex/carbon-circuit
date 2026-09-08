package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	evidencev1 "github.com/carboncircuit/backend/gen/carboncircuit/evidence/v1"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
	"github.com/carboncircuit/backend/services/api-gateway/internal/upstream"
)

const (
	EvidenceBodyLimit  = 25 * 1024 * 1024
	multipartSpillSize = 1 * 1024 * 1024
)

var evidencePurposeByName = map[string]evidencev1.DocumentPurpose{
	"claim_evidence":      evidencev1.DocumentPurpose_DOCUMENT_PURPOSE_CLAIM_EVIDENCE,
	"batch_certification": evidencev1.DocumentPurpose_DOCUMENT_PURPOSE_BATCH_CERTIFICATION,
}

var evidencePurposeName = invertMap(evidencePurposeByName)

var scanStatusName = map[evidencev1.ScanStatus]string{
	evidencev1.ScanStatus_SCAN_STATUS_PENDING: "pending",
	evidencev1.ScanStatus_SCAN_STATUS_CLEAN:   "clean",
	evidencev1.ScanStatus_SCAN_STATUS_FAILED:  "failed",
}

var scanVerdictName = map[evidencev1.ScanVerdict]string{
	evidencev1.ScanVerdict_SCAN_VERDICT_ACCEPTED:                  "accepted",
	evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_MEDIA_TYPE:       "rejected_media_type",
	evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_CONTENT_MISMATCH: "rejected_content_mismatch",
	evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_TOO_LARGE:        "rejected_too_large",
	evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_PAGE_COUNT:       "rejected_page_count",
	evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_ACTIVE_CONTENT:   "rejected_active_content",
	evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_MALWARE:          "rejected_malware",
	evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_UNREADABLE:       "rejected_unreadable",
}

type documentResponse struct {
	ID                    string `json:"id"`
	Purpose               string `json:"purpose"`
	FileName              string `json:"file_name"`
	MediaType             string `json:"media_type"`
	ByteSize              int64  `json:"byte_size"`
	PageCount             *int32 `json:"page_count"`
	ContentHash           string `json:"content_hash"`
	ScanStatus            string `json:"scan_status"`
	ScanVerdict           string `json:"scan_verdict"`
	ScannedBy             string `json:"scanned_by"`
	ScannedAt             string `json:"scanned_at"`
	ActiveContentStripped bool   `json:"active_content_stripped"`
	CreatedAt             string `json:"created_at"`
}

func toDocumentResponse(document *evidencev1.Document) documentResponse {
	var pages *int32
	if document.GetPageCount() > 0 {
		counted := document.GetPageCount()
		pages = &counted
	}

	return documentResponse{
		ID:                    document.GetId(),
		Purpose:               evidencePurposeName[document.GetPurpose()],
		FileName:              document.GetFileName(),
		MediaType:             document.GetMediaType(),
		ByteSize:              document.GetByteSize(),
		PageCount:             pages,
		ContentHash:           document.GetContentHash(),
		ScanStatus:            scanStatusName[document.GetScanStatus()],
		ScanVerdict:           scanVerdictName[document.GetScanVerdict()],
		ScannedBy:             document.GetScannedBy(),
		ScannedAt:             document.GetScannedAt(),
		ActiveContentStripped: document.GetActiveContentStripped(),
		CreatedAt:             document.GetCreatedAt(),
	}
}

func (h *Handlers) UploadEvidence(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	idempotencyKey, present := httpx.IdempotencyKeyFrom(c.Request.Context())
	if !present {
		httpx.Fail(c, httpx.CodeIdempotencyKeyRequired)
		return
	}

	if err := h.widenUploadWindow(c); err != nil {
		h.Logger.Warn("upload window could not be widened, a large file may time out",
			errorAttributes(c, err)...)
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, EvidenceBodyLimit)

	if err := c.Request.ParseMultipartForm(multipartSpillSize); err != nil {
		httpx.Fail(c, httpx.CodePayloadTooLarge)
		return
	}
	defer c.Request.MultipartForm.RemoveAll()

	purpose, known := evidencePurposeByName[c.Request.FormValue("purpose")]
	if !known {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "purpose", Code: "UNSUPPORTED_VALUE",
		})
		return
	}

	header, err := singleUpload(c)
	if err != nil {
		httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
			Field: "file", Code: "REQUIRED",
		})
		return
	}

	if header.Size > EvidenceBodyLimit {
		httpx.Fail(c, httpx.CodePayloadTooLarge)
		return
	}

	content, err := header.Open()
	if err != nil {
		httpx.Fail(c, httpx.CodeValidation)
		return
	}
	defer content.Close()

	uploaded, err := h.Evidence.UploadDocument(c.Request.Context(), upstream.Upload{
		Purpose:           purpose,
		FileName:          header.Filename,
		DeclaredMediaType: header.Header.Get("Content-Type"),
		IdempotencyKey:    idempotencyKey,
		Content:           content,
	})
	if err != nil {
		h.failEvidence(c, err)
		return
	}

	statusCode := http.StatusCreated
	if uploaded.GetAlreadyUploaded() {
		statusCode = http.StatusOK
	}

	httpx.Data(c, statusCode, toDocumentResponse(uploaded.GetDocument()))
}

func (h *Handlers) EvidenceResponseWindow() time.Duration {
	return h.EvidenceUploadWindow + h.EvidenceUploadTimeout
}

func (h *Handlers) widenUploadWindow(c *gin.Context) error {
	if h.EvidenceUploadWindow <= 0 {
		return nil
	}

	controller := http.NewResponseController(c.Writer)
	started := time.Now()

	if err := controller.SetReadDeadline(started.Add(h.EvidenceUploadWindow)); err != nil {
		return fmt.Errorf("set read deadline: %w", err)
	}

	if err := controller.SetWriteDeadline(started.Add(h.EvidenceResponseWindow())); err != nil {
		return fmt.Errorf("set write deadline: %w", err)
	}

	return nil
}

func singleUpload(c *gin.Context) (*multipart.FileHeader, error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return nil, err
	}
	file.Close()

	return header, nil
}

func (h *Handlers) GetEvidence(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	found, err := h.Evidence.GetDocument(c.Request.Context(), c.Param("documentId"))
	if err != nil {
		h.failEvidence(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, toDocumentResponse(found.GetDocument()))
}

func (h *Handlers) ListEvidence(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	purpose := evidencePurposeByName[c.Query("purpose")]

	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			httpx.Fail(c, httpx.CodeValidation, httpx.FieldError{
				Field: "limit", Code: "UNSUPPORTED_VALUE",
			})
			return
		}
		limit = parsed
	}

	listed, err := h.Evidence.ListDocuments(
		c.Request.Context(), purpose, c.Query("after"), int32(limit),
	)
	if err != nil {
		h.failEvidence(c, err)
		return
	}

	documents := make([]documentResponse, 0, len(listed.GetDocuments()))
	for _, document := range listed.GetDocuments() {
		documents = append(documents, toDocumentResponse(document))
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"documents":   documents,
		"next_cursor": emptyToNil(listed.GetNextCursor()),
	})
}

func (h *Handlers) CreateEvidenceDownloadLink(c *gin.Context) {
	if _, resolved := caller.ContextFrom(c.Request.Context()); !resolved {
		httpx.Fail(c, httpx.CodeUnauthenticated)
		return
	}

	link, err := h.Evidence.CreateDownloadLink(c.Request.Context(), c.Param("documentId"))
	if err != nil {
		h.failEvidence(c, err)
		return
	}

	httpx.Data(c, http.StatusOK, map[string]any{
		"url":        link.GetUrl(),
		"expires_at": link.GetExpiresAt(),
		"file_name":  link.GetFileName(),
		"media_type": link.GetMediaType(),
	})
}

func refusalReason(err error) string {
	for _, detail := range status.Convert(err).Details() {
		info, matches := detail.(*errdetails.ErrorInfo)
		if matches && info.GetDomain() == events.EvidenceRefusalDomain {
			return info.GetReason()
		}
	}
	return "REJECTED"
}

func (h *Handlers) failEvidence(c *gin.Context, err error) {
	switch status.Code(err) {
	case codes.NotFound:
		httpx.Fail(c, httpx.CodeResourceNotFound)
	case codes.PermissionDenied:
		httpx.Fail(c, httpx.CodeOrganizationReadOnly)
	case codes.InvalidArgument:
		httpx.Fail(c, httpx.CodeUnsupportedMediaType, httpx.FieldError{
			Field: "file", Code: refusalReason(err),
		})
	case codes.FailedPrecondition:
		httpx.Fail(c, httpx.CodeEvidenceNotReady)
	case codes.AlreadyExists:
		httpx.Fail(c, httpx.CodeIdempotencyKeyReused)
	case codes.Aborted:
		httpx.Fail(c, httpx.CodeRequestInProgress)
	default:
		h.Logger.Error("evidence call failed", errorAttributes(c, err)...)
		httpx.Fail(c, httpx.CodeDependencyUnavailable)
	}
}
