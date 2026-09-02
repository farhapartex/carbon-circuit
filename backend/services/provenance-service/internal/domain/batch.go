package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/domain"
)

type ProductCategory string

const (
	CategoryElectronics ProductCategory = "electronics"
	CategoryAgriculture ProductCategory = "agriculture"
	CategoryPharma      ProductCategory = "pharma"
	CategoryTextiles    ProductCategory = "textiles"
)

type CheckpointType string

const (
	ProductionComplete CheckpointType = "production_complete"
	DepartedOrigin     CheckpointType = "departed_origin"
	CustomsExport      CheckpointType = "customs_export"
	CustomsImport      CheckpointType = "customs_import"
	ArrivedDestination CheckpointType = "arrived_destination"
)

type AnchorStatus string

const (
	Unanchored  AnchorStatus = "unanchored"
	Provisional AnchorStatus = "provisional"
	Confirmed   AnchorStatus = "confirmed"
)

const (
	MaximumParentChainDepth = 10
	CorrectionWindowDays    = 7
	PublicReferenceLength   = 22
)

var ExpectedCheckpointSequence = map[ProductCategory][]CheckpointType{
	CategoryElectronics: {
		ProductionComplete,
		DepartedOrigin,
		CustomsExport,
		CustomsImport,
		ArrivedDestination,
	},
	CategoryAgriculture: {},
	CategoryPharma:      {},
	CategoryTextiles:    {},
}

type Batch struct {
	domain.Base
	OrganizationID          uuid.UUID             `gorm:"column:organization_id;type:uuid"`
	OriginatingFacilityID   uuid.UUID             `gorm:"column:originating_facility_id;type:uuid"`
	OriginatingFacilityName string                `gorm:"column:originating_facility_name"`
	PublicReference         string                `gorm:"column:public_reference;type:char(22)"`
	ProductCategory         ProductCategory       `gorm:"column:product_category"`
	ComponentType           string                `gorm:"column:component_type"`
	LotNumber               *string               `gorm:"column:lot_number"`
	Quantity                string                `gorm:"column:quantity;type:numeric(28,6)"`
	Unit                    string                `gorm:"column:unit"`
	ProducedAt              time.Time             `gorm:"column:produced_at"`
	ExternalID              *string               `gorm:"column:external_id"`
	CheckpointCount         int                   `gorm:"column:checkpoint_count"`
	ProvenanceScore         int                   `gorm:"column:provenance_score"`
	ScoreComponents         database.JSONDocument `gorm:"column:score_components;type:json"`
}

func (Batch) TableName() string { return "batches" }

type BatchParent struct {
	domain.Base
	OrganizationID    uuid.UUID  `gorm:"column:organization_id;type:uuid"`
	BatchID           uuid.UUID  `gorm:"column:batch_id;type:uuid"`
	DeclaredReference string     `gorm:"column:declared_reference;type:char(22)"`
	ParentBatchID     *uuid.UUID `gorm:"column:parent_batch_id;type:uuid"`
}

func (BatchParent) TableName() string { return "batch_parents" }

type Checkpoint struct {
	domain.Base
	OrganizationID             uuid.UUID      `gorm:"column:organization_id;type:uuid"`
	BatchID                    uuid.UUID      `gorm:"column:batch_id;type:uuid"`
	Type                       CheckpointType `gorm:"column:type"`
	LocationLabel              string         `gorm:"column:location_label"`
	CountryCode                string         `gorm:"column:country_code;type:char(2)"`
	Latitude                   *string        `gorm:"column:latitude;type:numeric(9,6)"`
	Longitude                  *string        `gorm:"column:longitude;type:numeric(9,6)"`
	ShippingMethod             *string        `gorm:"column:shipping_method"`
	OccurredAt                 time.Time      `gorm:"column:occurred_at"`
	ReportedAt                 time.Time      `gorm:"column:reported_at"`
	ReportedByOrganizationID   uuid.UUID      `gorm:"column:reported_by_organization_id;type:uuid"`
	ReportedByOrganizationName string         `gorm:"column:reported_by_organization_name"`
	AnchorStatus               AnchorStatus   `gorm:"column:anchor_status"`
	AnchorEpoch                *int           `gorm:"column:anchor_epoch"`
	AnchorTransactionHash      *string        `gorm:"column:anchor_transaction_hash;type:char(66)"`
	InclusionProofAvailable    bool           `gorm:"column:inclusion_proof_available"`
	SupersedesCheckpointID     *uuid.UUID     `gorm:"column:supersedes_checkpoint_id;type:uuid"`
	SupersededByCheckpointID   *uuid.UUID     `gorm:"column:superseded_by_checkpoint_id;type:uuid"`
	CorrectionReason           *string        `gorm:"column:correction_reason"`
	ExternalID                 *string        `gorm:"column:external_id"`
}

func (Checkpoint) TableName() string { return "checkpoints" }

func (c Checkpoint) MovesGoods() bool { return c.Type != ProductionComplete }
