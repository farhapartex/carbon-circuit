package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/carboncircuit/backend/internal/domain"
)

type OrganizationType string

const (
	OrganizationManufacturer OrganizationType = "manufacturer"
	OrganizationAssembler    OrganizationType = "assembler"
	OrganizationLogistics    OrganizationType = "logistics"
	OrganizationCreditBuyer  OrganizationType = "credit_buyer"
)

type OrganizationState string

const (
	OrganizationActive     OrganizationState = "active"
	OrganizationRestricted OrganizationState = "restricted"
	OrganizationReadOnly   OrganizationState = "read_only"
	OrganizationSuspended  OrganizationState = "suspended"
)

type VerificationStatus string

const (
	VerificationVerified   VerificationStatus = "verified"
	VerificationUnverified VerificationStatus = "unverified"
	VerificationRejected   VerificationStatus = "rejected"
)

type Organization struct {
	domain.Base
	Name                       string             `gorm:"column:name"`
	Type                       OrganizationType   `gorm:"column:type"`
	CountryOfIncorporation     string             `gorm:"column:country_of_incorporation;type:char(2)"`
	BusinessRegistrationNumber string             `gorm:"column:business_registration_number"`
	VerificationStatus         VerificationStatus `gorm:"column:verification_status"`
	State                      OrganizationState  `gorm:"column:state"`
	ProductCategories          pq.StringArray     `gorm:"column:product_categories;type:text[]"`
	RegistryRecordID           *uuid.UUID         `gorm:"column:registry_record_id;type:uuid"`
	NameSimilarity             *string            `gorm:"column:name_similarity;type:numeric(4,3)"`
	RejectionReason            *RegistryRejection `gorm:"column:rejection_reason"`
	VerifiedAt                 *time.Time         `gorm:"column:verified_at"`
}

func (Organization) TableName() string { return "organizations" }
