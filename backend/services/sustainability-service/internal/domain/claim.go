package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/domain"
)

type ActivityType string

const (
	RenewableEnergy          ActivityType = "renewable_energy"
	ReducedEmissionLogistics ActivityType = "reduced_emission_logistics"
	ResponsibleSourcing      ActivityType = "responsible_sourcing"
)

type ClaimStatus string

const (
	Submitted                ClaimStatus = "submitted"
	AIReview                 ClaimStatus = "ai_review"
	HumanReview              ClaimStatus = "human_review"
	Approved                 ClaimStatus = "approved"
	Rejected                 ClaimStatus = "rejected"
	MoreInformationRequested ClaimStatus = "more_information_requested"
)

type QueuePriority string

const (
	Normal   QueuePriority = "normal"
	High     QueuePriority = "high"
	Critical QueuePriority = "critical"
)

type FactorKind string

const (
	GridEmission      FactorKind = "grid_emission"
	LogisticsBaseline FactorKind = "logistics_baseline"
	MaterialAvoided   FactorKind = "material_avoided"
)

const (
	DualApprovalThreshold = "5000"
	MaximumEvidenceCount  = 25
)

type Claim struct {
	domain.Base
	OrganizationID        uuid.UUID             `gorm:"column:organization_id;type:uuid"`
	SubmittedByUserID     uuid.UUID             `gorm:"column:submitted_by_user_id;type:uuid"`
	FacilityID            uuid.UUID             `gorm:"column:facility_id;type:uuid"`
	FacilityName          string                `gorm:"column:facility_name"`
	ActivityType          ActivityType          `gorm:"column:activity_type"`
	VintageYear           int                   `gorm:"column:vintage_year"`
	PeriodStart           time.Time             `gorm:"column:period_start"`
	PeriodEnd             time.Time             `gorm:"column:period_end"`
	DeclaredFigures       database.JSONDocument `gorm:"column:declared_figures;type:jsonb"`
	RequestedAmount       string                `gorm:"column:requested_amount;type:numeric(28,6)"`
	ComputedCeiling       string                `gorm:"column:computed_ceiling;type:numeric(28,6)"`
	CapacityBasis         string                `gorm:"column:capacity_basis;type:numeric(28,6)"`
	CapacitySource        string                `gorm:"column:capacity_source"`
	DiscountFactor        string                `gorm:"column:discount_factor;type:numeric(3,2)"`
	ReferenceFactorID     uuid.UUID             `gorm:"column:reference_factor_id;type:uuid"`
	ReferenceFactorValue  string                `gorm:"column:reference_factor_value;type:numeric(20,6)"`
	Status                ClaimStatus           `gorm:"column:status"`
	Priority              QueuePriority         `gorm:"column:priority"`
	RequiresDualApproval  bool                  `gorm:"column:requires_dual_approval"`
	ExclusivityAttestedAt time.Time             `gorm:"column:exclusivity_attested_at"`
	ExclusivityAttestedBy uuid.UUID             `gorm:"column:exclusivity_attested_by"`
	IssuedAmount          *string               `gorm:"column:issued_amount;type:numeric(28,6)"`
}

func (Claim) TableName() string { return "claims" }

type ClaimEvidence struct {
	domain.Base
	OrganizationID uuid.UUID `gorm:"column:organization_id;type:uuid"`
	ClaimID        uuid.UUID `gorm:"column:claim_id;type:uuid"`
	EvidenceID     uuid.UUID `gorm:"column:evidence_id;type:uuid"`
	ContentHash    string    `gorm:"column:content_hash;type:char(64)"`
	FileName       string    `gorm:"column:file_name"`
	MediaType      string    `gorm:"column:media_type"`
	PageCount      *int      `gorm:"column:page_count"`
}

func (ClaimEvidence) TableName() string { return "claim_evidence" }

type ReferenceFactor struct {
	domain.Base
	Kind          FactorKind `gorm:"column:kind"`
	LookupKey     string     `gorm:"column:lookup_key"`
	Unit          string     `gorm:"column:unit"`
	Factor        string     `gorm:"column:factor;type:numeric(20,6)"`
	EffectiveFrom time.Time  `gorm:"column:effective_from"`
	EffectiveTo   *time.Time `gorm:"column:effective_to"`
}

func (ReferenceFactor) TableName() string { return "reference_factors" }
