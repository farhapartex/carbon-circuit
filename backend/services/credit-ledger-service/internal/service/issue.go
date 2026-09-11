package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/inbox"
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/creditclass"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/domain"
	"github.com/carboncircuit/backend/services/credit-ledger-service/internal/repository"
)

const issuanceAggregate = "credit_issuance"

var (
	ErrUnknownTopic      = errors.New("no handler is registered for this topic")
	ErrNoTreasury        = errors.New("the organization has no treasury address to issue into")
	ErrAmountNotPositive = errors.New("an issuance amount must be greater than zero")
)

var chainScale = decimal.New(1, 18)

type Facility struct {
	ID          uuid.UUID
	Name        string
	CountryCode string
}

type Treasury struct {
	Address string
}

type OrganizationResolver interface {
	Facility(ctx context.Context, facilityID uuid.UUID) (Facility, error)
	Treasury(ctx context.Context, organizationID uuid.UUID) (Treasury, error)
}

type Issuer struct {
	database      *gorm.DB
	ledger        repository.LedgerStore
	organizations OrganizationResolver
	consumerGroup string
	logger        *slog.Logger
}

func NewIssuer(
	handle *gorm.DB,
	ledger repository.LedgerStore,
	organizations OrganizationResolver,
	consumerGroup string,
	logger *slog.Logger,
) *Issuer {
	return &Issuer{
		database:      handle,
		ledger:        ledger,
		organizations: organizations,
		consumerGroup: consumerGroup,
		logger:        logger,
	}
}

func (i *Issuer) Apply(ctx context.Context, delivery kafka.Delivery) error {
	eventID, err := uuid.Parse(delivery.EventID)
	if err != nil {
		return fmt.Errorf("event on %q carries no usable event id", delivery.Topic)
	}

	if delivery.Topic != events.TopicClaimDecisionRecorded {
		return fmt.Errorf("%w: %s", ErrUnknownTopic, delivery.Topic)
	}

	var decided events.ClaimDecisionRecorded
	if err := json.Unmarshal(delivery.Value, &decided); err != nil {
		return fmt.Errorf("decode claim decision: %w", err)
	}

	if decided.Status != "approved" {
		return i.acknowledge(ctx, eventID, delivery)
	}

	return i.issue(ctx, eventID, delivery, decided)
}

func (i *Issuer) acknowledge(
	ctx context.Context,
	eventID uuid.UUID,
	delivery kafka.Delivery,
) error {
	return database.Within(ctx, i.database, func(tx database.Tx) error {
		err := inbox.Claim(tx, inbox.Receipt{
			EventID:       eventID,
			Topic:         delivery.Topic,
			ConsumerGroup: i.consumerGroup,
		})
		if errors.Is(err, inbox.ErrAlreadyProcessed) {
			return nil
		}
		return err
	})
}

func (i *Issuer) issue(
	ctx context.Context,
	eventID uuid.UUID,
	delivery kafka.Delivery,
	decided events.ClaimDecisionRecorded,
) error {
	organizationID, err := uuid.Parse(decided.OrganizationID)
	if err != nil {
		return fmt.Errorf("decision carries an unusable organization id")
	}

	claimID, err := uuid.Parse(decided.ClaimID)
	if err != nil {
		return fmt.Errorf("decision carries an unusable claim id")
	}

	facilityID, err := uuid.Parse(decided.FacilityID)
	if err != nil {
		return fmt.Errorf("decision carries an unusable facility id")
	}

	amount, err := decimal.NewFromString(decided.IssuedAmount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: %q", ErrAmountNotPositive, decided.IssuedAmount)
	}

	facility, err := i.organizations.Facility(ctx, facilityID)
	if err != nil {
		return err
	}

	treasury, err := i.organizations.Treasury(ctx, organizationID)
	if err != nil {
		return err
	}
	if treasury.Address == "" {
		return fmt.Errorf("%w: organization %s", ErrNoTreasury, organizationID)
	}

	tokenID, err := creditclass.TokenID(facilityID, decided.VintageYear, decided.ActivityType)
	if err != nil {
		return err
	}

	return database.Within(ctx, i.database, func(tx database.Tx) error {
		claimErr := inbox.Claim(tx, inbox.Receipt{
			EventID:       eventID,
			Topic:         delivery.Topic,
			ConsumerGroup: i.consumerGroup,
		})
		if errors.Is(claimErr, inbox.ErrAlreadyProcessed) {
			return nil
		}
		if claimErr != nil {
			return claimErr
		}

		class := domain.CreditClass{
			TokenID:         tokenID,
			FacilityID:      facilityID,
			FacilityName:    facility.Name,
			FacilityCountry: facility.CountryCode,
			VintageYear:     decided.VintageYear,
			ActivityType:    domain.ActivityType(decided.ActivityType),
		}

		if err := i.ledger.EnsureClass(tx, &class); err != nil {
			return err
		}

		if err := database.AdoptTenant(tx, database.TenantContext{
			OrganizationID: organizationID.String(),
		}); err != nil {
			return err
		}

		existing, found, err := i.ledger.IssuedForClaim(tx, claimID)
		if err != nil {
			return err
		}
		if found {
			i.logger.Info("claim already issued, leaving the ledger untouched",
				slog.String("claim_id", claimID.String()),
				slog.String("issuance_id", existing.ID.String()))
			return nil
		}

		issuance := domain.CreditIssuance{
			OrganizationID:  organizationID,
			CreditClassID:   class.ID,
			ClaimID:         claimID,
			Amount:          amount.StringFixed(domain.Places),
			TreasuryAddress: treasury.Address,
			AnchorState:     domain.Unanchored,
			IssuedAt:        time.Now().UTC(),
		}

		if err := i.ledger.RecordIssuance(tx, &issuance); err != nil {
			if errors.Is(err, repository.ErrClaimAlreadyIssued) {
				return nil
			}
			return err
		}

		if err := i.ledger.Credit(tx, organizationID, class.ID, issuance.Amount); err != nil {
			return err
		}

		return i.publish(tx, class, issuance, amount)
	})
}

func (i *Issuer) publish(
	tx database.Tx,
	class domain.CreditClass,
	issuance domain.CreditIssuance,
	amount decimal.Decimal,
) error {
	issuedAt := issuance.IssuedAt.Format(time.RFC3339)

	_, err := outbox.Append(tx, outbox.Envelope{
		AggregateType: issuanceAggregate,
		AggregateID:   issuance.ID,
		EventType:     events.TopicCreditIssued,
		Payload: events.CreditIssued{
			IssuanceID:      issuance.ID.String(),
			OrganizationID:  issuance.OrganizationID.String(),
			ClaimID:         issuance.ClaimID.String(),
			TokenID:         class.TokenID,
			FacilityID:      class.FacilityID.String(),
			FacilityName:    class.FacilityName,
			VintageYear:     class.VintageYear,
			ActivityType:    string(class.ActivityType),
			Amount:          issuance.Amount,
			TreasuryAddress: issuance.TreasuryAddress,
			AnchorState:     string(issuance.AnchorState),
			IssuedAt:        issuedAt,
		},
	})
	if err != nil {
		return err
	}

	_, err = outbox.Append(tx, outbox.Envelope{
		AggregateType: issuanceAggregate,
		AggregateID:   issuance.ID,
		EventType:     events.TopicChainMintRequested,
		Payload: events.ChainMintRequested{
			IssuanceID:      issuance.ID.String(),
			OrganizationID:  issuance.OrganizationID.String(),
			ClaimID:         issuance.ClaimID.String(),
			TokenID:         class.TokenID,
			TreasuryAddress: issuance.TreasuryAddress,
			AmountWei:       ChainAmount(amount),
			RequestedAt:     issuedAt,
		},
	})

	return err
}

func ChainAmount(amount decimal.Decimal) string {
	scaled := amount.Mul(chainScale)

	if !scaled.IsInteger() {
		return ""
	}

	return new(big.Int).Set(scaled.BigInt()).String()
}
