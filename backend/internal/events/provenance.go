package events

const (
	TopicBatchCreated     = "batch.created"
	TopicCheckpointLogged = "checkpoint.logged"
)

type ScoreComponent struct {
	Label       string `json:"label"`
	Earned      int    `json:"earned"`
	Available   int    `json:"available"`
	Explanation string `json:"explanation"`
}

type BatchCreated struct {
	BatchID                    string           `json:"batch_id"`
	OrganizationID             string           `json:"organization_id"`
	PublicReference            string           `json:"public_reference"`
	ProductCategory            string           `json:"product_category"`
	ComponentType              string           `json:"component_type"`
	OriginatingFacilityName    string           `json:"originating_facility_name"`
	OriginatingFacilityCountry string           `json:"originating_facility_country"`
	ProducedAt                 string           `json:"produced_at"`
	ProvenanceScore            int              `json:"provenance_score"`
	ScoreComponents            []ScoreComponent `json:"score_components"`
	OccurredAt                 string           `json:"occurred_at"`
}

type CheckpointLogged struct {
	BatchID                string           `json:"batch_id"`
	CheckpointID           string           `json:"checkpoint_id"`
	OrganizationID         string           `json:"organization_id"`
	Type                   string           `json:"type"`
	LocationLabel          string           `json:"location_label"`
	CountryCode            string           `json:"country_code"`
	ShippingMethod         string           `json:"shipping_method,omitempty"`
	OccurredAt             string           `json:"occurred_at"`
	AnchorStatus           string           `json:"anchor_status"`
	SupersedesCheckpointID string           `json:"supersedes_checkpoint_id,omitempty"`
	ProvenanceScore        int              `json:"provenance_score"`
	ScoreComponents        []ScoreComponent `json:"score_components"`
}
