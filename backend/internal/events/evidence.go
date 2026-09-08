package events

const TopicEvidenceScanned = "evidence.scanned"

type EvidenceScanned struct {
	EvidenceID       string `json:"evidence_id"`
	OrganizationID   string `json:"organization_id"`
	Purpose          string `json:"purpose"`
	ContentHash      string `json:"content_hash"`
	MediaType        string `json:"media_type"`
	ByteSize         int64  `json:"byte_size"`
	PageCount        int    `json:"page_count,omitempty"`
	ScanStatus       string `json:"scan_status"`
	ScanVerdict      string `json:"scan_verdict"`
	ScannedBy        string `json:"scanned_by"`
	ScannedAt        string `json:"scanned_at"`
	UploadedByUserID string `json:"uploaded_by_user_id"`
}
