package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/apikey"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
)

var (
	ErrAPIKeyRejected = errors.New("api key is not valid")
	ErrAPIKeyRevoked  = errors.New("api key has been revoked")
)

type ValidatedAPIKey struct {
	KeyID              uuid.UUID
	Prefix             string
	OrganizationID     uuid.UUID
	OrganizationName   string
	OrganizationType   domain.OrganizationType
	OrganizationState  domain.OrganizationState
	VerificationStatus domain.VerificationStatus
	ActingUserID       uuid.UUID
}

func (s *APIKeyService) Validate(
	ctx context.Context,
	presented string,
) (ValidatedAPIKey, error) {
	parsed, err := apikey.Parse(presented)
	if err != nil {
		return ValidatedAPIKey{}, ErrAPIKeyRejected
	}

	var key domain.APIKey

	err = database.Within(ctx, s.database, func(tx database.Tx) error {
		found, present, findErr := s.keys.FindByPrefix(tx, parsed.Prefix)
		if findErr != nil {
			return findErr
		}
		if !present {
			return ErrAPIKeyRejected
		}
		if !s.hasher.Matches(parsed.Secret, found.SecretHMAC) {
			return ErrAPIKeyRejected
		}
		if !found.Active() {
			return ErrAPIKeyRevoked
		}

		key = found

		return nil
	})
	if err != nil {
		return ValidatedAPIKey{}, err
	}

	var validated ValidatedAPIKey

	err = database.WithinTenant(ctx, s.database, database.TenantContext{
		UserID:         key.CreatedByUserID.String(),
		OrganizationID: key.OrganizationID.String(),
	}, func(tx database.Tx) error {
		organization, present, orgErr := s.organizations.Find(tx, key.OrganizationID)
		if orgErr != nil {
			return orgErr
		}
		if !present {
			return ErrAPIKeyRejected
		}

		if touchErr := s.keys.TouchLastUsed(tx, key.ID, time.Now().UTC()); touchErr != nil {
			return touchErr
		}

		validated = ValidatedAPIKey{
			KeyID:              key.ID,
			Prefix:             key.Prefix,
			OrganizationID:     organization.ID,
			OrganizationName:   organization.Name,
			OrganizationType:   organization.Type,
			OrganizationState:  organization.State,
			VerificationStatus: organization.VerificationStatus,
			ActingUserID:       key.CreatedByUserID,
		}

		return nil
	})
	if err != nil {
		return ValidatedAPIKey{}, err
	}

	return validated, nil
}
