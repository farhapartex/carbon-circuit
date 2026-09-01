package registry

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

type Declaration struct {
	LegalName          string
	CountryCode        string
	RegistrationNumber string
}

type Outcome struct {
	Status         domain.VerificationStatus
	State          domain.OrganizationState
	Rejection      *domain.RegistryRejection
	RecordID       *uuid.UUID
	MatchFound     bool
	NameSimilarity *float64
}

func (o Outcome) SimilarityString() string {
	if o.NameSimilarity == nil {
		return ""
	}
	return strconv.FormatFloat(*o.NameSimilarity, 'f', 3, 64)
}

func (o Outcome) SimilarityOrNil() *string {
	if o.NameSimilarity == nil {
		return nil
	}
	formatted := o.SimilarityString()
	return &formatted
}

func Unmatched() Outcome {
	return Outcome{
		Status: domain.VerificationUnverified,
		State:  domain.OrganizationActive,
	}
}

func Assess(declaration Declaration, record domain.BusinessRegistryRecord) Outcome {
	similarity := NameSimilarity(declaration.LegalName, record.LegalName)
	recordID := record.ID

	outcome := Outcome{
		MatchFound:     true,
		RecordID:       &recordID,
		NameSimilarity: &similarity,
	}

	switch rejection := rejectionFor(record, similarity); rejection {
	case nil:
		outcome.Status = domain.VerificationVerified
		outcome.State = domain.OrganizationActive
	default:
		outcome.Status = domain.VerificationRejected
		outcome.State = domain.OrganizationSuspended
		outcome.Rejection = rejection
	}

	return outcome
}

func rejectionFor(
	record domain.BusinessRegistryRecord,
	similarity float64,
) *domain.RegistryRejection {
	if record.EntityStatus == domain.RegistryDissolved {
		return pointerTo(domain.RejectionEntityDissolved)
	}
	if record.Sanctioned {
		return pointerTo(domain.RejectionSanctionsFlag)
	}
	if similarity < MinimumNameSimilarity {
		return pointerTo(domain.RejectionNameMismatch)
	}
	return nil
}

func pointerTo(rejection domain.RegistryRejection) *domain.RegistryRejection {
	return &rejection
}

func (d Declaration) Validate() error {
	if len(d.CountryCode) != 2 {
		return fmt.Errorf("country of incorporation must be a two letter code")
	}
	if d.RegistrationNumber == "" {
		return fmt.Errorf("business registration number is required")
	}
	if d.LegalName == "" {
		return fmt.Errorf("registered legal name is required")
	}
	return nil
}
