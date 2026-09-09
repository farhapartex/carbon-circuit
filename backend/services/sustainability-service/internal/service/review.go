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
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/inbox"
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/domain"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/repository"
)

var ErrUnknownTopic = errors.New("no handler is registered for this topic")

const StubReviewer = "stub:not-assessed"

type ReviewPipeline struct {
	database      *gorm.DB
	claims        repository.ClaimStore
	consumerGroup string
	stubEnabled   bool
	logger        *slog.Logger
}

func NewReviewPipeline(
	handle *gorm.DB,
	claims repository.ClaimStore,
	consumerGroup string,
	stubEnabled bool,
	logger *slog.Logger,
) *ReviewPipeline {
	if stubEnabled {
		logger.Warn("answering ai review requests with a stub",
			slog.String("consequence",
				"every claim reaches a verifier with no assessment, recorded as assessed_by="+StubReviewer))
	}

	return &ReviewPipeline{
		database:      handle,
		claims:        claims,
		consumerGroup: consumerGroup,
		stubEnabled:   stubEnabled,
		logger:        logger,
	}
}

func (p *ReviewPipeline) Apply(ctx context.Context, delivery kafka.Delivery) error {
	eventID, err := uuid.Parse(delivery.EventID)
	if err != nil {
		return fmt.Errorf("event on %q carries no usable event id", delivery.Topic)
	}

	return database.Within(ctx, p.database, func(tx database.Tx) error {
		claimErr := inbox.Claim(tx, inbox.Receipt{
			EventID:       eventID,
			Topic:         delivery.Topic,
			ConsumerGroup: p.consumerGroup,
		})
		if errors.Is(claimErr, inbox.ErrAlreadyProcessed) {
			return nil
		}
		if claimErr != nil {
			return claimErr
		}

		switch delivery.Topic {
		case events.TopicClaimAIReviewRequested:
			return p.answerWithStub(tx, delivery.Value)
		case events.TopicClaimAIReviewCompleted:
			return p.applyCompletedReview(tx, delivery.Value)
		default:
			return fmt.Errorf("%w: %s", ErrUnknownTopic, delivery.Topic)
		}
	})
}

func (p *ReviewPipeline) answerWithStub(tx database.Tx, payload []byte) error {
	if !p.stubEnabled {
		return nil
	}

	var requested events.ClaimAIReviewRequested
	if err := json.Unmarshal(payload, &requested); err != nil {
		return fmt.Errorf("decode ai review request: %w", err)
	}

	claimID, err := uuid.Parse(requested.ClaimID)
	if err != nil {
		return fmt.Errorf("ai review request carries an unusable claim id")
	}

	_, err = outbox.Append(tx, outbox.Envelope{
		AggregateType: claimAggregate,
		AggregateID:   claimID,
		EventType:     events.TopicClaimAIReviewCompleted,
		Payload: events.ClaimAIReviewCompleted{
			ClaimID:        requested.ClaimID,
			OrganizationID: requested.OrganizationID,
			Assessment:     string(domain.NotAssessed),
			Narrative: "No assessment was produced. The AI review service is not deployed, " +
				"so this claim reaches a verifier with the evidence alone.",
			AssessedBy: StubReviewer,
			AssessedAt: time.Now().UTC().Format(time.RFC3339),
		},
	})

	return err
}

func (p *ReviewPipeline) applyCompletedReview(tx database.Tx, payload []byte) error {
	var completed events.ClaimAIReviewCompleted
	if err := json.Unmarshal(payload, &completed); err != nil {
		return fmt.Errorf("decode completed ai review: %w", err)
	}

	claimID, err := uuid.Parse(completed.ClaimID)
	if err != nil {
		return fmt.Errorf("completed ai review carries an unusable claim id")
	}

	organizationID, err := uuid.Parse(completed.OrganizationID)
	if err != nil {
		return fmt.Errorf("completed ai review carries an unusable organization id")
	}

	scoped := database.TenantContext{OrganizationID: organizationID.String()}
	if err := database.AdoptTenant(tx, scoped); err != nil {
		return err
	}

	review := domain.ClaimAIReview{
		OrganizationID:   organizationID,
		ClaimID:          claimID,
		Assessment:       assessmentOf(completed.Assessment),
		ExtractedFigures: encode(completed.ExtractedFigures),
		Flags:            encode(completed.Flags),
		Citations:        database.JSONDocument("[]"),
		Narrative:        completed.Narrative,
		AssessedBy:       completed.AssessedBy,
		AssessedAt:       parseTime(completed.AssessedAt),
	}

	if completed.Confidence != "" {
		confidence := completed.Confidence
		review.Confidence = &confidence
	}

	if err := p.claims.RecordAIReview(tx, &review); err != nil {
		return err
	}

	advanced, err := p.claims.AdvanceStatus(
		tx, organizationID, claimID, domain.Submitted, domain.HumanReview,
	)
	if err != nil {
		return err
	}

	if !advanced {
		p.logger.Info("claim was not waiting on an ai review",
			slog.String("claim_id", claimID.String()))
	}

	return nil
}

func assessmentOf(value string) domain.AIAssessment {
	switch domain.AIAssessment(value) {
	case domain.Corroborated,
		domain.CorroboratedWithDiscrepancy,
		domain.Uncorroborated,
		domain.Contradicted:
		return domain.AIAssessment(value)
	default:
		return domain.NotAssessed
	}
}

func encode(value any) database.JSONDocument {
	if value == nil {
		return database.JSONDocument("{}")
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return database.JSONDocument("{}")
	}

	return database.JSONDocument(encoded)
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Now().UTC()
	}
	return parsed
}
