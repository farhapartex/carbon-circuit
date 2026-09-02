package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/registry"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

var ErrOrganizationNotFound = errors.New("organization not found")

type Detail struct {
	Organization       domain.Organization
	TreasuryDesignated bool
	TreasuryAddress    string
	Outcome            registry.Outcome
}

type OrganizationReader struct {
	database      *gorm.DB
	organizations repository.OrganizationReader
	logger        *slog.Logger
}

func NewOrganizationReader(
	handle *gorm.DB,
	organizations repository.OrganizationReader,
	logger *slog.Logger,
) *OrganizationReader {
	return &OrganizationReader{database: handle, organizations: organizations, logger: logger}
}

func (r *OrganizationReader) Detail(
	ctx context.Context,
	organizationID uuid.UUID,
) (Detail, error) {
	var detail Detail

	err := database.WithinTenant(
		ctx,
		r.database,
		database.TenantContext{OrganizationID: organizationID.String()},
		func(tx database.Tx) error {
			organization, found, err := r.organizations.Find(tx, organizationID)
			if err != nil {
				return err
			}
			if !found {
				return ErrOrganizationNotFound
			}

			treasuryAddress, designated, err := r.organizations.ActiveTreasuryAddress(tx, organizationID)
			if err != nil {
				return err
			}

			outcome := outcomeOf(organization)

			if organization.RegistryRecordID != nil {
				record, found, recordErr := r.organizations.FindRecordByID(tx, *organization.RegistryRecordID)
				if recordErr != nil {
					return recordErr
				}
				if found {
					outcome.RegisteredName = record.LegalName
				}
			}

			detail = Detail{
				Organization:       organization,
				TreasuryDesignated: designated,
				TreasuryAddress:    treasuryAddress,
				Outcome:            outcome,
			}
			return nil
		},
	)
	if err != nil {
		return Detail{}, err
	}

	return detail, nil
}

func outcomeOf(organization domain.Organization) registry.Outcome {
	outcome := registry.Outcome{
		Status:     organization.VerificationStatus,
		State:      organization.State,
		Rejection:  organization.RejectionReason,
		MatchFound: organization.RegistryRecordID != nil,
		RecordID:   organization.RegistryRecordID,
	}

	if organization.NameSimilarity != nil {
		outcome.NameSimilarity = parseSimilarity(*organization.NameSimilarity)
	}

	return outcome
}

func parseSimilarity(stored string) *float64 {
	var value float64
	if _, err := fmt.Sscanf(stored, "%f", &value); err != nil {
		return nil
	}
	return &value
}
