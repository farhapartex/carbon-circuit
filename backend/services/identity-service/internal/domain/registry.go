package domain

import (
	"time"

	"github.com/lib/pq"

	"github.com/carboncircuit/backend/internal/domain"
)

type RegistryEntityStatus string

const (
	RegistryActive    RegistryEntityStatus = "active"
	RegistryDissolved RegistryEntityStatus = "dissolved"
)

type RegistryRejection string

const (
	RejectionEntityDissolved RegistryRejection = "entity_dissolved"
	RejectionSanctionsFlag   RegistryRejection = "sanctions_flag"
	RejectionNameMismatch    RegistryRejection = "name_mismatch"
)

type BusinessRegistryRecord struct {
	domain.Base
	CountryCode        string               `gorm:"column:country_code;type:char(2)"`
	RegistrationNumber string               `gorm:"column:registration_number"`
	LegalName          string               `gorm:"column:legal_name"`
	RegisteredAddress  string               `gorm:"column:registered_address"`
	IncorporationDate  time.Time            `gorm:"column:incorporation_date"`
	EntityStatus       RegistryEntityStatus `gorm:"column:entity_status"`
	IndustryCodes      pq.StringArray       `gorm:"column:industry_codes;type:text[]"`
	Sanctioned         bool                 `gorm:"column:sanctioned"`
}

func (BusinessRegistryRecord) TableName() string { return "business_registry_records" }
