package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

var ErrNotResolvable = errors.New("no such record")

type ResolvedOrganization struct {
	ID                     uuid.UUID
	Name                   string
	CountryOfIncorporation string
	TreasuryAddress        string
}

type ResolvedFacility struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	CountryCode    string
}

type ResolutionService struct {
	database *gorm.DB
	records  repository.ResolutionStore
	logger   *slog.Logger
}

func NewResolutionService(
	handle *gorm.DB,
	records repository.ResolutionStore,
	logger *slog.Logger,
) *ResolutionService {
	return &ResolutionService{database: handle, records: records, logger: logger}
}

func (s *ResolutionService) Organization(
	ctx context.Context,
	organizationID uuid.UUID,
) (ResolvedOrganization, error) {
	var resolved ResolvedOrganization

	err := database.Within(ctx, s.database, func(tx database.Tx) error {
		organization, found, err := s.records.Organization(tx, organizationID)
		if err != nil {
			return err
		}
		if !found {
			return ErrNotResolvable
		}

		address, err := s.records.TreasuryAddress(tx, organizationID)
		if err != nil {
			return err
		}

		resolved = ResolvedOrganization{
			ID:                     organization.ID,
			Name:                   organization.Name,
			CountryOfIncorporation: organization.CountryOfIncorporation,
			TreasuryAddress:        address,
		}
		return nil
	})
	if err != nil {
		return ResolvedOrganization{}, err
	}

	return resolved, nil
}

func (s *ResolutionService) Facility(
	ctx context.Context,
	facilityID uuid.UUID,
) (ResolvedFacility, error) {
	var resolved ResolvedFacility

	err := database.Within(ctx, s.database, func(tx database.Tx) error {
		facility, found, err := s.records.Facility(tx, facilityID)
		if err != nil {
			return err
		}
		if !found {
			return ErrNotResolvable
		}

		resolved = ResolvedFacility{
			ID:             facility.ID,
			OrganizationID: facility.OrganizationID,
			Name:           facility.Name,
			CountryCode:    facility.CountryCode,
		}
		return nil
	})
	if err != nil {
		return ResolvedFacility{}, err
	}

	return resolved, nil
}
