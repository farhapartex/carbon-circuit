package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/idempotency"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/wallet"
)

const designateTreasuryEndpoint = "POST /v1/treasury"

var (
	ErrNotOrganizationOwner = errors.New("only an organization owner may designate the treasury address")
	ErrTreasuryDesignated   = errors.New("organization already has an active treasury address")
	ErrAddressTaken         = errors.New("address is already the treasury of another organization")
	ErrProofRejected        = errors.New("ownership proof was rejected")
)

type Nonce struct {
	Value     string
	Domain    string
	ChainID   int
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type Designation struct {
	Address      string
	DesignatedAt time.Time
	Replayed     bool
}

type Ownership struct {
	Message        string
	Signature      string
	IdempotencyKey string
}

type TreasuryService struct {
	database    *gorm.DB
	treasury    repository.TreasuryStore
	expectation wallet.Expectation
	nonceWindow time.Duration
	logger      *slog.Logger
}

func NewTreasuryService(
	handle *gorm.DB,
	treasury repository.TreasuryStore,
	expectation wallet.Expectation,
	nonceWindow time.Duration,
	logger *slog.Logger,
) *TreasuryService {
	return &TreasuryService{
		database:    handle,
		treasury:    treasury,
		expectation: expectation,
		nonceWindow: nonceWindow,
		logger:      logger,
	}
}

func (s *TreasuryService) IssueNonce(ctx context.Context, userID uuid.UUID) (Nonce, error) {
	value, err := randomNonce()
	if err != nil {
		return Nonce{}, err
	}

	issued := time.Now()
	record := domain.SiweNonce{
		Nonce:     value,
		Domain:    s.expectation.Domain,
		UserID:    &userID,
		IssuedAt:  issued,
		ExpiresAt: issued.Add(s.nonceWindow),
	}

	err = database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{UserID: userID.String()},
		func(tx database.Tx) error {
			return s.treasury.IssueNonce(tx, &record)
		},
	)
	if err != nil {
		return Nonce{}, err
	}

	return Nonce{
		Value:     record.Nonce,
		Domain:    record.Domain,
		ChainID:   s.expectation.ChainID,
		IssuedAt:  record.IssuedAt,
		ExpiresAt: record.ExpiresAt,
	}, nil
}

func (s *TreasuryService) Designate(
	ctx context.Context,
	organizationID, userID uuid.UUID,
	role domain.OrganizationRole,
	ownership Ownership,
) (Designation, error) {
	if role != domain.RoleOwner {
		return Designation{}, ErrNotOrganizationOwner
	}

	proof, err := wallet.Verify(s.expectation, ownership.Message, ownership.Signature)
	if err != nil {
		s.logger.Warn("treasury ownership proof rejected",
			slog.String("organization_id", organizationID.String()),
			slog.Any("error", err),
		)
		return Designation{}, ErrProofRejected
	}

	var designation Designation

	err = database.WithinTenant(
		ctx,
		s.database,
		database.TenantContext{UserID: userID.String(), OrganizationID: organizationID.String()},
		func(tx database.Tx) error {
			result, replayed, workErr := s.persist(tx, organizationID, userID, proof, ownership)
			if workErr != nil {
				return workErr
			}
			designation = result
			designation.Replayed = replayed
			return nil
		},
	)
	if err != nil {
		return Designation{}, err
	}

	return designation, nil
}

func (s *TreasuryService) persist(
	tx database.Tx,
	organizationID, userID uuid.UUID,
	proof wallet.Proof,
	ownership Ownership,
) (Designation, bool, error) {
	reservation, err := idempotency.Reserve(tx, idempotency.Request{
		Scope:    idempotency.ForOrganization(organizationID),
		Endpoint: designateTreasuryEndpoint,
		Key:      ownership.IdempotencyKey,
		Body:     []byte(proof.Address + "\x1f" + proof.Nonce),
	})
	switch {
	case errors.Is(err, idempotency.ErrInProgress):
		return Designation{}, false, ErrRequestInProgress
	case errors.Is(err, idempotency.ErrKeyReused):
		return Designation{}, false, ErrIdempotencyConflict
	case err != nil:
		return Designation{}, false, err
	}

	if reservation.IsReplay() {
		return Designation{
			Address:      proof.Address,
			DesignatedAt: time.Now(),
		}, true, nil
	}

	existing, err := s.treasury.HasActiveTreasury(tx, organizationID)
	if err != nil {
		return Designation{}, false, err
	}
	if existing {
		return Designation{}, false, ErrTreasuryDesignated
	}

	consumed, err := s.treasury.ConsumeNonce(tx, proof.Nonce, time.Now())
	if err != nil {
		if errors.Is(err, repository.ErrNonceUnknown) {
			return Designation{}, false, ErrProofRejected
		}
		return Designation{}, false, err
	}

	if consumed.UserID == nil || *consumed.UserID != userID {
		return Designation{}, false, ErrProofRejected
	}

	treasury := domain.TreasuryAddress{
		OrganizationID:     organizationID,
		Address:            proof.Address,
		State:              domain.TreasuryActive,
		DesignatedByUserID: userID,
	}

	if err := s.treasury.InsertTreasury(tx, &treasury, consumed.ID, ownership.Signature); err != nil {
		if errors.Is(err, repository.ErrAddressDesignated) {
			return Designation{}, false, ErrAddressTaken
		}
		return Designation{}, false, err
	}

	designated := Designation{Address: treasury.Address, DesignatedAt: treasury.CreatedAt}

	if err := idempotency.Complete(tx, reservation.RecordID, idempotency.Response{
		Status:     201,
		Body:       []byte(fmt.Sprintf(`{"address":%q}`, treasury.Address)),
		ResourceID: &treasury.ID,
	}); err != nil {
		return Designation{}, false, err
	}

	return designated, false, nil
}

func randomNonce() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return hex.EncodeToString(raw), nil
}
