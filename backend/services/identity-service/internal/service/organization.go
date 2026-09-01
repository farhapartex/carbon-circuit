package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/registry"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
)

const (
	createOrganizationEndpoint = "POST /v1/organizations"
	organizationAggregate      = "organization"
	organizationVerifiedEvent  = "organization.verified"
)

var (
	ErrRegistrationTaken   = errors.New("registration number already registered")
	ErrOrganizationExists  = errors.New("user already belongs to an organization")
	ErrRequestInProgress   = errors.New("an identical request is already being processed")
	ErrIdempotencyConflict = errors.New("idempotency key was used with a different request")
)

type Registration struct {
	LegalName          string
	Type               domain.OrganizationType
	CountryCode        string
	RegistrationNumber string
	ProductCategories  []string
	IdempotencyKey     string
	RequestBody        []byte
}

type Registered struct {
	Organization domain.Organization
	Role         domain.OrganizationRole
	Outcome      registry.Outcome
	Replayed     bool
}

type OrganizationService struct {
	database      *gorm.DB
	users         repository.UserStore
	memberships   repository.MembershipStore
	organizations repository.OrganizationWriter
	lookup        repository.RegistryLookup
	logger        *slog.Logger
}

func NewOrganizationService(
	handle *gorm.DB,
	users repository.UserStore,
	memberships repository.MembershipStore,
	organizations repository.OrganizationWriter,
	lookup repository.RegistryLookup,
	logger *slog.Logger,
) *OrganizationService {
	return &OrganizationService{
		database:      handle,
		users:         users,
		memberships:   memberships,
		organizations: organizations,
		lookup:        lookup,
		logger:        logger,
	}
}

func (s *OrganizationService) Create(
	ctx context.Context,
	subject string,
	registration Registration,
) (Registered, error) {
	declaration := registry.Declaration{
		LegalName:          registration.LegalName,
		CountryCode:        registration.CountryCode,
		RegistrationNumber: registration.RegistrationNumber,
	}
	if err := declaration.Validate(); err != nil {
		return Registered{}, err
	}

	user, err := s.users.FindByAuth0Subject(ctx, subject)
	if err != nil {
		return Registered{}, err
	}

	organizationID, err := uuid.NewV7()
	if err != nil {
		return Registered{}, fmt.Errorf("generate organization id: %w", err)
	}

	var registered Registered

	err = database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{UserID: user.ID.String(), OrganizationID: organizationID.String()},
		func(tx database.Tx) error {
			outcome, replay, workErr := s.persist(tx, user, organizationID, registration, declaration)
			if workErr != nil {
				return workErr
			}
			registered = outcome
			registered.Replayed = replay
			return nil
		},
	)
	if err != nil {
		return Registered{}, err
	}

	return registered, nil
}

func (s *OrganizationService) persist(
	tx database.Tx,
	user domain.User,
	organizationID uuid.UUID,
	registration Registration,
	declaration registry.Declaration,
) (Registered, bool, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForUser(user.ID),
		Endpoint: createOrganizationEndpoint,
		Key:      registration.IdempotencyKey,
		Body:     registration.RequestBody,
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return Registered{}, false, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return Registered{}, false, ErrIdempotencyConflict
	case err != nil:
		return Registered{}, false, err
	}

	if reservation.IsReplay() {
		replayed, decodeErr := decodeReplay(reservation.Replay.Body)
		if decodeErr != nil {
			return Registered{}, false, decodeErr
		}
		return replayed, true, nil
	}

	enrolled, err := s.organizations.HasActiveMembership(tx, user.ID)
	if err != nil {
		return Registered{}, false, err
	}
	if enrolled {
		return Registered{}, false, ErrOrganizationExists
	}

	outcome := registry.Unmatched()

	record, found, err := s.lookup.FindRecord(tx, declaration.CountryCode, declaration.RegistrationNumber)
	if err != nil {
		return Registered{}, false, err
	}
	if found {
		outcome = registry.Assess(declaration, record)
	}

	organization := domain.Organization{
		Name:                       registration.LegalName,
		Type:                       registration.Type,
		CountryOfIncorporation:     declaration.CountryCode,
		BusinessRegistrationNumber: declaration.RegistrationNumber,
		VerificationStatus:         outcome.Status,
		State:                      outcome.State,
		ProductCategories:          pq.StringArray(registration.ProductCategories),
		RegistryRecordID:           outcome.RecordID,
		NameSimilarity:             outcome.SimilarityOrNil(),
		RejectionReason:            outcome.Rejection,
		VerifiedAt:                 verifiedTimestamp(outcome),
	}
	organization.ID = organizationID

	if err := s.organizations.Insert(tx, &organization); err != nil {
		return Registered{}, false, err
	}

	joinedAt := time.Now()
	membership := domain.OrganizationMembership{
		OrganizationID: organization.ID,
		UserID:         user.ID,
		Role:           domain.RoleOwner,
		State:          domain.MembershipActive,
		JoinedAt:       &joinedAt,
	}

	if err := s.organizations.InsertMembership(tx, &membership); err != nil {
		return Registered{}, false, err
	}

	if outcome.Status == domain.VerificationVerified {
		if _, err := outbox.Append(tx, outbox.Envelope{
			AggregateType: organizationAggregate,
			AggregateID:   organization.ID,
			EventType:     organizationVerifiedEvent,
			Payload: map[string]string{
				"organization_id":   organization.ID.String(),
				"organization_type": string(organization.Type),
			},
		}); err != nil {
			return Registered{}, false, err
		}
	}

	registered := Registered{
		Organization: organization,
		Role:         domain.RoleOwner,
		Outcome:      outcome,
	}

	body, err := json.Marshal(replayOf(registered))
	if err != nil {
		return Registered{}, false, fmt.Errorf("encode idempotent response: %w", err)
	}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     201,
		Body:       body,
		ResourceID: &organization.ID,
	}); err != nil {
		return Registered{}, false, err
	}

	return registered, false, nil
}

func verifiedTimestamp(outcome registry.Outcome) *time.Time {
	if outcome.Status != domain.VerificationVerified {
		return nil
	}
	now := time.Now()
	return &now
}
