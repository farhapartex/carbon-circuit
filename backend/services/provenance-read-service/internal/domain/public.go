package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/domain"
)

type PublicBatch struct {
	domain.Base
	PublicReference            string                `gorm:"column:public_reference;type:char(22)"`
	ProductCategory            string                `gorm:"column:product_category"`
	ComponentType              string                `gorm:"column:component_type"`
	OriginatingFacilityName    string                `gorm:"column:originating_facility_name"`
	OriginatingFacilityCountry string                `gorm:"column:originating_facility_country;type:char(2)"`
	ProducedAt                 time.Time             `gorm:"column:produced_at"`
	Projected                  bool                  `gorm:"column:projected"`
	ProvenanceScore            int                   `gorm:"column:provenance_score"`
	ScoreComponents            database.JSONDocument `gorm:"column:score_components;type:json"`
	LastUpdatedAt              time.Time             `gorm:"column:last_updated_at"`
}

func (PublicBatch) TableName() string { return "public_batches" }

type PublicCheckpoint struct {
	domain.Base
	BatchID               uuid.UUID `gorm:"column:batch_id;type:uuid"`
	Type                  string    `gorm:"column:type"`
	LocationLabel         string    `gorm:"column:location_label"`
	CountryCode           string    `gorm:"column:country_code;type:char(2)"`
	ShippingMethod        *string   `gorm:"column:shipping_method"`
	OccurredAt            time.Time `gorm:"column:occurred_at"`
	AnchorStatus          string    `gorm:"column:anchor_status"`
	AnchorEpoch           *int      `gorm:"column:anchor_epoch"`
	AnchorTransactionHash *string   `gorm:"column:anchor_transaction_hash"`
	Superseded            bool      `gorm:"column:superseded"`
}

func (PublicCheckpoint) TableName() string { return "public_checkpoints" }
