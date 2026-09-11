package events

const (
	TopicCreditIssued       = "credit.issued"
	TopicCreditRetired      = "credit.retired"
	TopicChainMintRequested = "chain.mint.requested"

	CreditRefusalDomain = "credit-ledger-service"
)

type CreditIssued struct {
	IssuanceID      string `json:"issuance_id"`
	OrganizationID  string `json:"organization_id"`
	ClaimID         string `json:"claim_id"`
	TokenID         string `json:"token_id"`
	FacilityID      string `json:"facility_id"`
	FacilityName    string `json:"facility_name"`
	VintageYear     int    `json:"vintage_year"`
	ActivityType    string `json:"activity_type"`
	Amount          string `json:"amount"`
	TreasuryAddress string `json:"treasury_address"`
	AnchorState     string `json:"anchor_state"`
	IssuedAt        string `json:"issued_at"`
}

type ChainMintRequested struct {
	IssuanceID      string `json:"issuance_id"`
	OrganizationID  string `json:"organization_id"`
	ClaimID         string `json:"claim_id"`
	TokenID         string `json:"token_id"`
	TreasuryAddress string `json:"treasury_address"`
	AmountWei       string `json:"amount_wei"`
	RequestedAt     string `json:"requested_at"`
}
