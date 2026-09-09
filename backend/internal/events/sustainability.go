package events

const (
	TopicClaimSubmitted         = "claim.submitted"
	TopicClaimAIReviewRequested = "claim.ai_review.requested"
	TopicClaimAIReviewCompleted = "claim.ai_review.completed"
	TopicClaimDecisionRecorded  = "claim.decision.recorded"

	ClaimRefusalDomain = "sustainability-service"
)

type ClaimSubmitted struct {
	ClaimID              string   `json:"claim_id"`
	OrganizationID       string   `json:"organization_id"`
	FacilityID           string   `json:"facility_id"`
	ActivityType         string   `json:"activity_type"`
	VintageYear          int      `json:"vintage_year"`
	PeriodStart          string   `json:"period_start"`
	PeriodEnd            string   `json:"period_end"`
	RequestedAmount      string   `json:"requested_amount"`
	ComputedCeiling      string   `json:"computed_ceiling"`
	Priority             string   `json:"priority"`
	RequiresDualApproval bool     `json:"requires_dual_approval"`
	EvidenceHashes       []string `json:"evidence_hashes"`
	SubmittedAt          string   `json:"submitted_at"`
}

type ClaimAIReviewRequested struct {
	ClaimID        string   `json:"claim_id"`
	OrganizationID string   `json:"organization_id"`
	ActivityType   string   `json:"activity_type"`
	VintageYear    int      `json:"vintage_year"`
	EvidenceIDs    []string `json:"evidence_ids"`
	EvidenceHashes []string `json:"evidence_hashes"`
	RequestedAt    string   `json:"requested_at"`
}

type ClaimAIReviewCompleted struct {
	ClaimID          string            `json:"claim_id"`
	OrganizationID   string            `json:"organization_id"`
	Assessment       string            `json:"assessment"`
	Confidence       string            `json:"confidence,omitempty"`
	ExtractedFigures map[string]string `json:"extracted_figures,omitempty"`
	Flags            []string          `json:"flags,omitempty"`
	Narrative        string            `json:"narrative"`
	AssessedBy       string            `json:"assessed_by"`
	AssessedAt       string            `json:"assessed_at"`
}
