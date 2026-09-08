package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type Purpose string

const (
	ClaimEvidence      Purpose = "claim_evidence"
	BatchCertification Purpose = "batch_certification"
)

type ScanStatus string

const (
	ScanPending ScanStatus = "pending"
	ScanClean   ScanStatus = "clean"
	ScanFailed  ScanStatus = "failed"
)

type ScanVerdict string

const (
	Accepted                ScanVerdict = "accepted"
	RejectedMediaType       ScanVerdict = "rejected_media_type"
	RejectedContentMismatch ScanVerdict = "rejected_content_mismatch"
	RejectedTooLarge        ScanVerdict = "rejected_too_large"
	RejectedPageCount       ScanVerdict = "rejected_page_count"
	RejectedActiveContent   ScanVerdict = "rejected_active_content"
	RejectedMalware         ScanVerdict = "rejected_malware"
	RejectedUnreadable      ScanVerdict = "rejected_unreadable"
)

const (
	MaximumByteSize  = 25 * 1024 * 1024
	MaximumPageCount = 100
)

var AcceptedMediaTypes = map[string]struct{}{
	"application/pdf": {},
	"image/png":       {},
	"image/jpeg":      {},
	"text/csv":        {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {},
}

func MediaTypeAccepted(mediaType string) bool {
	_, accepted := AcceptedMediaTypes[mediaType]
	return accepted
}

type Document struct {
	domain.Base
	OrganizationID        uuid.UUID   `gorm:"column:organization_id;type:uuid"`
	UploadedByUserID      uuid.UUID   `gorm:"column:uploaded_by_user_id;type:uuid"`
	Purpose               Purpose     `gorm:"column:purpose"`
	FileName              string      `gorm:"column:file_name"`
	DeclaredMediaType     string      `gorm:"column:declared_media_type"`
	DetectedMediaType     string      `gorm:"column:detected_media_type"`
	ByteSize              int64       `gorm:"column:byte_size"`
	PageCount             *int        `gorm:"column:page_count"`
	ContentHash           string      `gorm:"column:content_hash;type:char(64)"`
	StoredHash            *string     `gorm:"column:stored_hash;type:char(64)"`
	StorageKey            *string     `gorm:"column:storage_key"`
	ScanStatus            ScanStatus  `gorm:"column:scan_status"`
	ScanVerdict           ScanVerdict `gorm:"column:scan_verdict"`
	ScannedBy             string      `gorm:"column:scanned_by"`
	ScannedAt             time.Time   `gorm:"column:scanned_at"`
	ActiveContentStripped bool        `gorm:"column:active_content_stripped"`
}

func (Document) TableName() string { return "documents" }

func (d Document) Usable() bool {
	return d.ScanStatus == ScanClean && d.StorageKey != nil && d.DeletedAt.Time.IsZero()
}
