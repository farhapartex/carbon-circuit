package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/registry"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

const createFacilityEndpoint = "POST /v1/facilities"

var (
	ErrFacilityNotFound     = errors.New("facility not found")
	ErrOrganizationRequired = errors.New("an organization is required")
)

var (
	gridRegions = map[string]struct{}{
		"US-CAISO": {}, "US-ERCOT": {}, "US-PJM": {}, "US-MISO": {},
		"EU-DE": {}, "EU-FR": {}, "EU-PL": {}, "UK": {},
		"CN-East": {}, "CN-South": {}, "IN-North": {}, "JP": {},
		"KR": {}, "TW": {}, "VN": {}, "MY": {}, "SG": {}, "TH": {},
	}

	facilityTypes = map[domain.FacilityType]struct{}{
		domain.FacilityRawMaterialProcessing: {},
		domain.FacilityComponentFabrication:  {},
		domain.FacilityAssembly:              {},
		domain.FacilityDistribution:          {},
	}

	plainDecimal = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)
)

type FacilityDeclaration struct {
	Name              string
	Address           string
	CountryCode       string
	GridRegion        string
	Type              domain.FacilityType
	FacilityReference string
	DeclaredCapacity  string
	DeclaredEnergyKwh string
	IdempotencyKey    string
	RequestBody       []byte
}

func (d FacilityDeclaration) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("facility name is required")
	}
	if d.Address == "" {
		return fmt.Errorf("facility address is required")
	}
	if len(d.CountryCode) != 2 {
		return fmt.Errorf("country must be a two letter code")
	}
	if _, ok := gridRegions[d.GridRegion]; !ok {
		return fmt.Errorf("grid region %q is not a known region", d.GridRegion)
	}
	if _, ok := facilityTypes[d.Type]; !ok {
		return fmt.Errorf("facility type %q is not a known type", d.Type)
	}
	if err := validateFigure("declared capacity", d.DeclaredCapacity); err != nil {
		return err
	}
	return validateFigure("declared energy consumption", d.DeclaredEnergyKwh)
}

func validateFigure(label, value string) error {
	if !plainDecimal.MatchString(value) {
		return fmt.Errorf("%s must be a decimal with up to six places", label)
	}
	if isZeroFigure(value) {
		return fmt.Errorf("%s must be greater than zero", label)
	}
	return nil
}

func isZeroFigure(value string) bool {
	for _, character := range value {
		if character >= '1' && character <= '9' {
			return false
		}
	}
	return true
}

type RegisteredFacility struct {
	Facility domain.Facility
	Replayed bool
}

type FacilityService struct {
	database      *gorm.DB
	facilities    repository.FacilityStore
	organizations repository.OrganizationReader
	logger        *slog.Logger
}

func NewFacilityService(
	handle *gorm.DB,
	facilities repository.FacilityStore,
	organizations repository.OrganizationReader,
	logger *slog.Logger,
) *FacilityService {
	return &FacilityService{
		database:      handle,
		facilities:    facilities,
		organizations: organizations,
		logger:        logger,
	}
}

func (s *FacilityService) Create(
	ctx context.Context,
	actor Actor,
	declaration FacilityDeclaration,
) (RegisteredFacility, error) {
	if !actor.manages() {
		return RegisteredFacility{}, ErrNotPermitted
	}
	if err := declaration.Validate(); err != nil {
		return RegisteredFacility{}, err
	}

	facilityID, err := uuid.NewV7()
	if err != nil {
		return RegisteredFacility{}, fmt.Errorf("generate facility id: %w", err)
	}

	var registered RegisteredFacility

	err = database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         actor.UserID.String(),
			OrganizationID: actor.OrganizationID.String(),
		},
		func(tx database.Tx) error {
			created, replay, workErr := s.persist(
				tx, actor.OrganizationID, facilityID, declaration,
			)
			if workErr != nil {
				return workErr
			}
			registered = RegisteredFacility{Facility: created, Replayed: replay}
			return nil
		},
	)
	if err != nil {
		return RegisteredFacility{}, err
	}

	return registered, nil
}

func (s *FacilityService) persist(
	tx database.Tx,
	organizationID, facilityID uuid.UUID,
	declaration FacilityDeclaration,
) (domain.Facility, bool, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForOrganization(organizationID),
		Endpoint: createFacilityEndpoint,
		Key:      declaration.IdempotencyKey,
		Body:     declaration.RequestBody,
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return domain.Facility{}, false, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return domain.Facility{}, false, ErrIdempotencyConflict
	case err != nil:
		return domain.Facility{}, false, err
	}

	if reservation.IsReplay() {
		var replayed domain.Facility
		if err := json.Unmarshal(reservation.Replay.Body, &replayed); err != nil {
			return domain.Facility{}, false, fmt.Errorf("decode replayed facility: %w", err)
		}
		return replayed, true, nil
	}

	organization, found, err := s.organizations.Find(tx, organizationID)
	if err != nil {
		return domain.Facility{}, false, err
	}
	if !found {
		return domain.Facility{}, false, ErrOrganizationRequired
	}

	outcome, err := s.assess(tx, organization, declaration)
	if err != nil {
		return domain.Facility{}, false, err
	}

	facility := domain.Facility{
		OrganizationID:        organizationID,
		Name:                  declaration.Name,
		Address:               declaration.Address,
		CountryCode:           declaration.CountryCode,
		GridRegion:            declaration.GridRegion,
		Type:                  declaration.Type,
		FacilityReference:     referenceOrNil(declaration.FacilityReference),
		VerificationStatus:    outcome.Status,
		CeilingDiscountFactor: outcome.DiscountFactor,
		TrustTier:             domain.TrustNew,
		DeclaredCapacity:      declaration.DeclaredCapacity,
		DeclaredEnergyKwh:     declaration.DeclaredEnergyKwh,
		AttestedCapacity:      outcome.AttestedCapacity,
		AttestedEnergyKwh:     outcome.AttestedEnergyKwh,
	}
	facility.ID = facilityID

	if err := s.facilities.Insert(tx, &facility); err != nil {
		return domain.Facility{}, false, err
	}

	body, err := json.Marshal(facility)
	if err != nil {
		return domain.Facility{}, false, fmt.Errorf("encode idempotent response: %w", err)
	}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     201,
		Body:       body,
		ResourceID: &facility.ID,
	}); err != nil {
		return domain.Facility{}, false, err
	}

	return facility, false, nil
}

func (s *FacilityService) assess(
	tx database.Tx,
	organization domain.Organization,
	declaration FacilityDeclaration,
) (registry.FacilityOutcome, error) {
	if declaration.FacilityReference == "" {
		return registry.AssessUnmatchedFacility(organization.VerificationStatus), nil
	}

	record, found, err := s.facilities.FindRegistryRecord(
		tx,
		organization.BusinessRegistrationNumber,
		declaration.FacilityReference,
	)
	if err != nil {
		return registry.FacilityOutcome{}, err
	}
	if !found {
		return registry.AssessUnmatchedFacility(organization.VerificationStatus), nil
	}

	return registry.AssessFacilityMatch(record), nil
}

func referenceOrNil(reference string) *string {
	if reference == "" {
		return nil
	}
	return &reference
}

func (s *FacilityService) List(
	ctx context.Context,
	actor Actor,
) ([]domain.Facility, error) {
	var facilities []domain.Facility

	err := database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         actor.UserID.String(),
			OrganizationID: actor.OrganizationID.String(),
		},
		func(tx database.Tx) error {
			listed, err := s.facilities.List(tx, actor.OrganizationID)
			if err != nil {
				return err
			}
			facilities = listed
			return nil
		},
	)

	return facilities, err
}

func (s *FacilityService) Get(
	ctx context.Context,
	actor Actor,
	facilityID uuid.UUID,
) (domain.Facility, error) {
	var facility domain.Facility

	err := database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{
			UserID:         actor.UserID.String(),
			OrganizationID: actor.OrganizationID.String(),
		},
		func(tx database.Tx) error {
			found, exists, err := s.facilities.Find(tx, actor.OrganizationID, facilityID)
			if err != nil {
				return err
			}
			if !exists {
				return ErrFacilityNotFound
			}
			facility = found
			return nil
		},
	)
	if err != nil {
		return domain.Facility{}, err
	}

	return facility, nil
}
