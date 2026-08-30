package domain

import (
	"github.com/lib/pq"

	"github.com/carboncircuit/backend/internal/domain"
)

type PlanTier string

const (
	TierBuyer      PlanTier = "buyer"
	TierStarter    PlanTier = "starter"
	TierGrowth     PlanTier = "growth"
	TierEnterprise PlanTier = "enterprise"
)

type OrganizationType string

const (
	OrganizationManufacturer OrganizationType = "manufacturer"
	OrganizationAssembler    OrganizationType = "assembler"
	OrganizationLogistics    OrganizationType = "logistics"
	OrganizationCreditBuyer  OrganizationType = "credit_buyer"
)

type UsageDimension string

const (
	DimensionBatches     UsageDimension = "batches"
	DimensionCheckpoints UsageDimension = "checkpoints"
	DimensionFacilities  UsageDimension = "facilities"
	DimensionUsers       UsageDimension = "users"
	DimensionAIReviews   UsageDimension = "ai_reviews"
	DimensionStorageGB   UsageDimension = "storage_gb"
)

type PlanLimit struct {
	domain.Base
	PlanID             string         `gorm:"column:plan_id;type:uuid" json:"plan_id"`
	Dimension          UsageDimension `gorm:"column:dimension" json:"dimension"`
	Included           *int64         `gorm:"column:included" json:"included"`
	FairUseCeiling     *int64         `gorm:"column:fair_use_ceiling" json:"fair_use_ceiling"`
	OverageRateUSD     *string        `gorm:"column:overage_rate_usd;type:numeric(10,2)" json:"overage_rate_usd"`
	BlocksOnExhaustion bool           `gorm:"column:blocks_on_exhaustion" json:"blocks_on_exhaustion"`
}

func (PlanLimit) TableName() string { return "plan_limits" }

type Plan struct {
	domain.Base
	Tier                     PlanTier       `gorm:"column:tier" json:"tier"`
	Name                     string         `gorm:"column:name" json:"name"`
	Audience                 string         `gorm:"column:audience" json:"audience"`
	MonthlyPriceUSD          string         `gorm:"column:monthly_price_usd;type:numeric(10,2)" json:"monthly_price_usd"`
	PriceNote                *string        `gorm:"column:price_note" json:"price_note"`
	AllowedOrganizationTypes pq.StringArray `gorm:"column:allowed_organization_types;type:text[]" json:"allowed_organization_types"`
	EvidenceStorageGB        *int32         `gorm:"column:evidence_storage_gb" json:"evidence_storage_gb"`
	PortalRatePerMinute      int32          `gorm:"column:portal_rate_per_minute" json:"portal_rate_per_minute"`
	APIRatePerMinute         *int32         `gorm:"column:api_rate_per_minute" json:"api_rate_per_minute"`
	APIKeyLimit              *int32         `gorm:"column:api_key_limit" json:"api_key_limit"`
	MarketplaceFeeBps        *int32         `gorm:"column:marketplace_fee_bps" json:"marketplace_fee_bps"`
	ReviewTurnaround         string         `gorm:"column:review_turnaround" json:"review_turnaround"`
	SupportLevel             string         `gorm:"column:support_level" json:"support_level"`
	Limits                   []PlanLimit    `gorm:"foreignKey:PlanID;references:ID" json:"limits"`
}

func (Plan) TableName() string { return "plans" }

func (p Plan) AllowsOrganizationType(organizationType OrganizationType) bool {
	for _, allowed := range p.AllowedOrganizationTypes {
		if OrganizationType(allowed) == organizationType {
			return true
		}
	}
	return false
}
