package service

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/registry"
)

type replayPayload struct {
	OrganizationID     uuid.UUID                 `json:"organization_id"`
	Name               string                    `json:"name"`
	Type               domain.OrganizationType   `json:"type"`
	CountryCode        string                    `json:"country_code"`
	RegistrationNumber string                    `json:"registration_number"`
	VerificationStatus domain.VerificationStatus `json:"verification_status"`
	State              domain.OrganizationState  `json:"state"`
	Role               domain.OrganizationRole   `json:"role"`
	Rejection          *domain.RegistryRejection `json:"rejection"`
	MatchFound         bool                      `json:"match_found"`
	NameSimilarity     *string                   `json:"name_similarity"`
}

func replayOf(registered Registered) replayPayload {
	organization := registered.Organization

	return replayPayload{
		OrganizationID:     organization.ID,
		Name:               organization.Name,
		Type:               organization.Type,
		CountryCode:        organization.CountryOfIncorporation,
		RegistrationNumber: organization.BusinessRegistrationNumber,
		VerificationStatus: organization.VerificationStatus,
		State:              organization.State,
		Role:               registered.Role,
		Rejection:          registered.Outcome.Rejection,
		MatchFound:         registered.Outcome.MatchFound,
		NameSimilarity:     registered.Outcome.SimilarityOrNil(),
	}
}

func decodeReplay(body []byte) (Registered, error) {
	var payload replayPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return Registered{}, fmt.Errorf("decode replayed organization: %w", err)
	}

	organization := domain.Organization{
		Name:                       payload.Name,
		Type:                       payload.Type,
		CountryOfIncorporation:     payload.CountryCode,
		BusinessRegistrationNumber: payload.RegistrationNumber,
		VerificationStatus:         payload.VerificationStatus,
		State:                      payload.State,
		RejectionReason:            payload.Rejection,
		NameSimilarity:             payload.NameSimilarity,
	}
	organization.ID = payload.OrganizationID

	return Registered{
		Organization: organization,
		Role:         payload.Role,
		Outcome: registry.Outcome{
			Status:     payload.VerificationStatus,
			State:      payload.State,
			Rejection:  payload.Rejection,
			MatchFound: payload.MatchFound,
		},
	}, nil
}
