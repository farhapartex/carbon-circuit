package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/services/billing-service/internal/domain"
	"github.com/carboncircuit/backend/services/billing-service/internal/repository"
)

const createSubscriptionEndpoint = "POST /v1/subscriptions"

var (
	ErrSubscriptionExists  = errors.New("organization already has a subscription")
	ErrPlanNotAllowed      = errors.New("plan is not available to this organization type")
	ErrRequestInProgress   = errors.New("an identical request is already being processed")
	ErrIdempotencyConflict = errors.New("idempotency key was used with a different request")
)

type Enrolment struct {
	OrganizationID   uuid.UUID
	OrganizationType domain.OrganizationType
	Tier             domain.PlanTier
	IdempotencyKey   string
	RequestBody      []byte
}

type Enrolled struct {
	Subscription domain.Subscription
	Replayed     bool
}

type SubscriptionCreator struct {
	database      *gorm.DB
	subscriptions repository.SubscriptionWriter
	logger        *slog.Logger
}

func NewSubscriptionCreator(
	handle *gorm.DB,
	subscriptions repository.SubscriptionWriter,
	logger *slog.Logger,
) *SubscriptionCreator {
	return &SubscriptionCreator{database: handle, subscriptions: subscriptions, logger: logger}
}

func (s *SubscriptionCreator) Create(ctx context.Context, enrolment Enrolment) (Enrolled, error) {
	var enrolled Enrolled

	err := database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{OrganizationID: enrolment.OrganizationID.String()},
		func(tx database.Tx) error {
			result, replayed, workErr := s.persist(tx, enrolment)
			if workErr != nil {
				return workErr
			}
			enrolled = result
			enrolled.Replayed = replayed
			return nil
		},
	)
	if err != nil {
		return Enrolled{}, err
	}

	return enrolled, nil
}

func (s *SubscriptionCreator) persist(
	tx database.Tx,
	enrolment Enrolment,
) (Enrolled, bool, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForOrganization(enrolment.OrganizationID),
		Endpoint: createSubscriptionEndpoint,
		Key:      enrolment.IdempotencyKey,
		Body:     enrolment.RequestBody,
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return Enrolled{}, false, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return Enrolled{}, false, ErrIdempotencyConflict
	case err != nil:
		return Enrolled{}, false, err
	}

	if reservation.IsReplay() {
		replayed, decodeErr := decodeSubscription(reservation.Replay.Body)
		if decodeErr != nil {
			return Enrolled{}, false, decodeErr
		}
		return Enrolled{Subscription: replayed}, true, nil
	}

	enrolled, err := s.enrol(tx, enrolment)
	if err != nil {
		return Enrolled{}, false, err
	}

	body, err := json.Marshal(enrolled.Subscription)
	if err != nil {
		return Enrolled{}, false, fmt.Errorf("encode idempotent response: %w", err)
	}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     201,
		Body:       body,
		ResourceID: &enrolled.Subscription.ID,
	}); err != nil {
		return Enrolled{}, false, err
	}

	return enrolled, false, nil
}

func (s *SubscriptionCreator) enrol(
	tx database.Tx,
	enrolment Enrolment,
) (Enrolled, error) {
	existing, err := s.subscriptions.HasSubscription(tx, enrolment.OrganizationID)
	if err != nil {
		return Enrolled{}, err
	}
	if existing {
		return Enrolled{}, ErrSubscriptionExists
	}

	plan, err := s.subscriptions.FindCurrentPlan(tx, enrolment.Tier)
	if err != nil {
		return Enrolled{}, err
	}

	if !plan.AllowsOrganizationType(enrolment.OrganizationType) {
		return Enrolled{}, ErrPlanNotAllowed
	}

	start := time.Now()
	subscription := domain.Subscription{
		OrganizationID:     enrolment.OrganizationID,
		PlanID:             plan.ID,
		State:              domain.SubscriptionActive,
		CurrentPeriodStart: start,
		CurrentPeriodEnd:   start.AddDate(0, 1, 0),
		Plan:               plan,
	}

	if err := s.subscriptions.Insert(tx, &subscription); err != nil {
		if errors.Is(err, repository.ErrSubscriptionExists) {
			return Enrolled{}, ErrSubscriptionExists
		}
		return Enrolled{}, err
	}

	return Enrolled{Subscription: subscription}, nil
}

func decodeSubscription(body []byte) (domain.Subscription, error) {
	var subscription domain.Subscription
	if err := json.Unmarshal(body, &subscription); err != nil {
		return domain.Subscription{}, fmt.Errorf("decode replayed subscription: %w", err)
	}
	return subscription, nil
}
