package domain

import (
	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/domain"
)

type FacilityType string

const (
	FacilityRawMaterialProcessing FacilityType = "raw_material_processing"
	FacilityComponentFabrication  FacilityType = "component_fabrication"
	FacilityAssembly              FacilityType = "assembly"
	FacilityDistribution          FacilityType = "distribution"
)

type FacilityVerification string

const (
	FacilityMatched     FacilityVerification = "facility_matched"
	OrganizationMatched FacilityVerification = "organization_matched"
	SelfDeclared        FacilityVerification = "self_declared"
)

type TrustTier string

const (
	TrustNew      TrustTier = "new"
	TrustVerified TrustTier = "verified"
	TrustTrusted  TrustTier = "trusted"
)

type Facility struct {
	domain.Base
	OrganizationID        uuid.UUID            `gorm:"column:organization_id;type:uuid"`
	Name                  string               `gorm:"column:name"`
	Address               string               `gorm:"column:address"`
	CountryCode           string               `gorm:"column:country_code;type:char(2)"`
	GridRegion            string               `gorm:"column:grid_region"`
	Type                  FacilityType         `gorm:"column:type"`
	FacilityReference     *string              `gorm:"column:facility_reference"`
	VerificationStatus    FacilityVerification `gorm:"column:verification_status"`
	CeilingDiscountFactor string               `gorm:"column:ceiling_discount_factor;type:numeric(3,2)"`
	TrustTier             TrustTier            `gorm:"column:trust_tier"`
	DeclaredCapacity      string               `gorm:"column:declared_capacity;type:numeric(20,6)"`
	DeclaredEnergyKwh     string               `gorm:"column:declared_energy_kwh;type:numeric(20,6)"`
	AttestedCapacity      *string              `gorm:"column:attested_capacity;type:numeric(20,6)"`
	AttestedEnergyKwh     *string              `gorm:"column:attested_energy_kwh;type:numeric(20,6)"`
}

func (Facility) TableName() string { return "facilities" }

type FacilityRegistryRecord struct {
	domain.Base
	OrganizationRegistrationNumber string `gorm:"column:organization_registration_number"`
	FacilityReference              string `gorm:"column:facility_reference"`
	AttestedCapacity               string `gorm:"column:attested_capacity;type:numeric(20,6)"`
	AttestedEnergyKwh              string `gorm:"column:attested_energy_kwh;type:numeric(20,6)"`
}

func (FacilityRegistryRecord) TableName() string { return "facility_registry_records" }
