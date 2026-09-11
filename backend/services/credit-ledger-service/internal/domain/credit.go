package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type ActivityType string

const (
	RenewableEnergy          ActivityType = "renewable_energy"
	ReducedEmissionLogistics ActivityType = "reduced_emission_logistics"
	ResponsibleSourcing      ActivityType = "responsible_sourcing"
)

type AnchorState string

const (
	Unanchored AnchorState = "unanchored"
	Submitted  AnchorState = "submitted"
	Anchored   AnchorState = "anchored"
)

const Places = 6

type CreditClass struct {
	domain.Base
	TokenID         string       `gorm:"column:token_id;type:numeric(78,0)"`
	FacilityID      uuid.UUID    `gorm:"column:facility_id;type:uuid"`
	FacilityName    string       `gorm:"column:facility_name"`
	FacilityCountry string       `gorm:"column:facility_country;type:char(2)"`
	VintageYear     int          `gorm:"column:vintage_year"`
	ActivityType    ActivityType `gorm:"column:activity_type"`
}

func (CreditClass) TableName() string { return "credit_classes" }

type CreditBalance struct {
	domain.Base
	OrganizationID uuid.UUID `gorm:"column:organization_id;type:uuid"`
	CreditClassID  uuid.UUID `gorm:"column:credit_class_id;type:uuid"`
	Available      string    `gorm:"column:available;type:numeric(28,6)"`
	Escrowed       string    `gorm:"column:escrowed;type:numeric(28,6)"`
	Retired        string    `gorm:"column:retired;type:numeric(28,6)"`
}

func (CreditBalance) TableName() string { return "credit_balances" }

type CreditIssuance struct {
	domain.Base
	OrganizationID  uuid.UUID   `gorm:"column:organization_id;type:uuid"`
	CreditClassID   uuid.UUID   `gorm:"column:credit_class_id;type:uuid"`
	ClaimID         uuid.UUID   `gorm:"column:claim_id;type:uuid"`
	Amount          string      `gorm:"column:amount;type:numeric(28,6)"`
	TreasuryAddress string      `gorm:"column:treasury_address;type:char(42)"`
	AnchorState     AnchorState `gorm:"column:anchor_state"`
	TransactionHash *string     `gorm:"column:transaction_hash;type:char(66)"`
	IssuedAt        time.Time   `gorm:"column:issued_at"`
	AnchoredAt      *time.Time  `gorm:"column:anchored_at"`
}

func (CreditIssuance) TableName() string { return "credit_issuances" }

type Holding struct {
	Class     CreditClass
	Available string
	Escrowed  string
	Retired   string
}
