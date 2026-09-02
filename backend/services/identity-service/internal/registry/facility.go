package registry

import (
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

const (
	DiscountFacilityMatched     = "1.00"
	DiscountOrganizationMatched = "0.75"
	DiscountSelfDeclared        = "0.50"
)

type FacilityOutcome struct {
	Status            domain.FacilityVerification
	DiscountFactor    string
	AttestedCapacity  *string
	AttestedEnergyKwh *string
}

func AssessFacilityMatch(record domain.FacilityRegistryRecord) FacilityOutcome {
	capacity := record.AttestedCapacity
	energy := record.AttestedEnergyKwh

	return FacilityOutcome{
		Status:            domain.FacilityMatched,
		DiscountFactor:    DiscountFacilityMatched,
		AttestedCapacity:  &capacity,
		AttestedEnergyKwh: &energy,
	}
}

func AssessUnmatchedFacility(
	organizationStatus domain.VerificationStatus,
) FacilityOutcome {
	if organizationStatus == domain.VerificationVerified {
		return FacilityOutcome{
			Status:         domain.OrganizationMatched,
			DiscountFactor: DiscountOrganizationMatched,
		}
	}

	return FacilityOutcome{
		Status:         domain.SelfDeclared,
		DiscountFactor: DiscountSelfDeclared,
	}
}
