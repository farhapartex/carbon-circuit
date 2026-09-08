package events

const (
	TopicClaimSubmitted         = "claim.submitted"
	TopicClaimAIReviewRequested = "claim.ai_review.requested"
	TopicClaimDecisionRecorded  = "claim.decision.recorded"
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
