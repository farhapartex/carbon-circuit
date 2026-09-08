package rpc

import (
	"time"

	evidencev1 "github.com/carboncircuit/backend/gen/carboncircuit/evidence/v1"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
	"github.com/carboncircuit/backend/services/evidence-service/internal/service"
)

var purposeNames = map[evidencev1.DocumentPurpose]domain.Purpose{
	evidencev1.DocumentPurpose_DOCUMENT_PURPOSE_CLAIM_EVIDENCE:      domain.ClaimEvidence,
	evidencev1.DocumentPurpose_DOCUMENT_PURPOSE_BATCH_CERTIFICATION: domain.BatchCertification,
}

var purposeCodes = map[domain.Purpose]evidencev1.DocumentPurpose{
	domain.ClaimEvidence:      evidencev1.DocumentPurpose_DOCUMENT_PURPOSE_CLAIM_EVIDENCE,
	domain.BatchCertification: evidencev1.DocumentPurpose_DOCUMENT_PURPOSE_BATCH_CERTIFICATION,
}

var statusCodes = map[domain.ScanStatus]evidencev1.ScanStatus{
	domain.ScanPending: evidencev1.ScanStatus_SCAN_STATUS_PENDING,
	domain.ScanClean:   evidencev1.ScanStatus_SCAN_STATUS_CLEAN,
	domain.ScanFailed:  evidencev1.ScanStatus_SCAN_STATUS_FAILED,
}

var verdictCodes = map[domain.ScanVerdict]evidencev1.ScanVerdict{
	domain.Accepted:                evidencev1.ScanVerdict_SCAN_VERDICT_ACCEPTED,
	domain.RejectedMediaType:       evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_MEDIA_TYPE,
	domain.RejectedContentMismatch: evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_CONTENT_MISMATCH,
	domain.RejectedTooLarge:        evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_TOO_LARGE,
	domain.RejectedPageCount:       evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_PAGE_COUNT,
	domain.RejectedActiveContent:   evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_ACTIVE_CONTENT,
	domain.RejectedMalware:         evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_MALWARE,
	domain.RejectedUnreadable:      evidencev1.ScanVerdict_SCAN_VERDICT_REJECTED_UNREADABLE,
}

func purposeFrom(code evidencev1.DocumentPurpose) (domain.Purpose, bool) {
	purpose, known := purposeNames[code]
	return purpose, known
}

func timestamp(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	return at.UTC().Format(time.RFC3339)
}

func documentMessage(document domain.Document) *evidencev1.Document {
	pages := 0
	if document.PageCount != nil {
		pages = *document.PageCount
	}

	return &evidencev1.Document{
		Id:                    document.ID.String(),
		OrganizationId:        document.OrganizationID.String(),
		UploadedByUserId:      document.UploadedByUserID.String(),
		Purpose:               purposeCodes[document.Purpose],
		FileName:              document.FileName,
		MediaType:             document.DetectedMediaType,
		ByteSize:              document.ByteSize,
		PageCount:             int32(pages),
		ContentHash:           document.ContentHash,
		ScanStatus:            statusCodes[document.ScanStatus],
		ScanVerdict:           verdictCodes[document.ScanVerdict],
		ScannedBy:             document.ScannedBy,
		ScannedAt:             timestamp(document.ScannedAt),
		ActiveContentStripped: document.ActiveContentStripped,
		CreatedAt:             timestamp(document.CreatedAt),
	}
}

func resolvedMessage(resolution service.Resolution) *evidencev1.ResolvedDocument {
	message := &evidencev1.ResolvedDocument{
		DocumentId: resolution.DocumentID.String(),
		Usable:     resolution.Usable,
		Refusal:    resolution.Refusal,
	}

	if !resolution.Usable {
		return message
	}

	pages := 0
	if resolution.Document.PageCount != nil {
		pages = *resolution.Document.PageCount
	}

	message.ContentHash = resolution.Document.ContentHash
	message.PageCount = int32(pages)
	message.MediaType = resolution.Document.DetectedMediaType
	message.ByteSize = resolution.Document.ByteSize

	return message
}
