package service_test

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/spruceid/siwe-go"
	"gorm.io/gorm"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/services/identity-service/internal/domain"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
	"github.com/carboncircuit/backend/services/identity-service/internal/wallet"
)

const (
	treasuryDomain = "localhost:3000"
	treasuryChain  = 31337
)

func treasuryService(handle *gorm.DB) *service.TreasuryService {
	return service.NewTreasuryService(
		handle,
		repository.NewTreasuryRepository(),
		wallet.Expectation{Domain: treasuryDomain, ChainID: treasuryChain},
		5*time.Minute,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func walletKey(t *testing.T) (*ecdsa.PrivateKey, string) {
	t.Helper()

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate wallet key: %v", err)
	}
	return key, crypto.PubkeyToAddress(key.PublicKey).Hex()
}

func signedProof(t *testing.T, key *ecdsa.PrivateKey, address, nonce string) (string, string) {
	t.Helper()

	built, err := siwe.InitMessage(treasuryDomain, address, "http://"+treasuryDomain, nonce, map[string]any{
		"chainId":   treasuryChain,
		"statement": "Designate this address as your CarbonCircuit Treasury Address.",
		"issuedAt":  time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("build siwe message: %v", err)
	}

	raw := built.String()
	hash := accounts.TextHash([]byte(raw))
	signature, err := crypto.Sign(hash, key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	signature[64] += 27

	return raw, fmt.Sprintf("0x%x", signature)
}

func seedOrganization(t *testing.T, handle *gorm.DB, userID uuid.UUID) uuid.UUID {
	t.Helper()

	organizationID := uuid.New()

	err := handle.Exec(
		`INSERT INTO identity.organizations
		   (id, name, type, country_of_incorporation, business_registration_number, verification_status)
		 VALUES (?, ?, 'manufacturer', 'TW', ?, 'unverified')`,
		organizationID, "Probe Org "+organizationID.String()[:8], "PROBE-"+organizationID.String()[:8],
	).Error
	if err != nil {
		t.Fatalf("seed organization: %v", err)
	}

	t.Cleanup(func() {
		scoped := database.TenantContext{
			UserID:         userID.String(),
			OrganizationID: organizationID.String(),
		}
		cleanupErr := database.WithinTenant(context.Background(), handle, scoped,
			func(tx database.Tx) error {
				tx.Session().Exec(`DELETE FROM identity.treasury_addresses WHERE organization_id = ?`, organizationID)
				tx.Session().Exec(`DELETE FROM identity.idempotency_records WHERE organization_id = ? OR user_id = ?`, organizationID, userID)
				return tx.Session().Exec(`UPDATE identity.organizations SET deleted_at = now() WHERE id = ?`, organizationID).Error
			})
		if cleanupErr != nil {
			t.Errorf("clean organization %s: %v", organizationID, cleanupErr)
		}
	})

	return organizationID
}

func ownership(message, signature, key string) service.Ownership {
	return service.Ownership{Message: message, Signature: signature, IdempotencyKey: key}
}

func TestOwnerDesignatesTheirTreasuryAddress(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID := seedOrganization(t, handle, userID)

	treasury := treasuryService(handle)
	nonce, err := treasury.IssueNonce(context.Background(), userID)
	if err != nil {
		t.Fatalf("issue nonce: %v", err)
	}

	key, address := walletKey(t)
	message, signature := signedProof(t, key, address, nonce.Value)

	designated, err := treasury.Designate(context.Background(), organizationID, userID,
		domain.RoleOwner, ownership(message, signature, "treasury-key-1"))
	if err != nil {
		t.Fatalf("designate: %v", err)
	}

	if designated.Address != lowerHex(address) {
		t.Fatalf("expected %s, got %s", lowerHex(address), designated.Address)
	}
}

func TestNonOwnerCannotDesignate(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID := seedOrganization(t, handle, userID)

	treasury := treasuryService(handle)
	nonce, _ := treasury.IssueNonce(context.Background(), userID)
	key, address := walletKey(t)
	message, signature := signedProof(t, key, address, nonce.Value)

	for _, role := range []domain.OrganizationRole{domain.RoleAdmin, domain.RoleMember} {
		t.Run(string(role), func(t *testing.T) {
			_, err := treasury.Designate(context.Background(), organizationID, userID,
				role, ownership(message, signature, "treasury-key-"+string(role)))
			if !errors.Is(err, service.ErrNotOrganizationOwner) {
				t.Fatalf("expected ErrNotOrganizationOwner, got %v", err)
			}
		})
	}
}

func TestNonceIsSingleUse(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	firstOrganization := seedOrganization(t, handle, userID)
	secondOrganization := seedOrganization(t, handle, userID)

	treasury := treasuryService(handle)
	nonce, _ := treasury.IssueNonce(context.Background(), userID)
	key, address := walletKey(t)
	message, signature := signedProof(t, key, address, nonce.Value)

	if _, err := treasury.Designate(context.Background(), firstOrganization, userID,
		domain.RoleOwner, ownership(message, signature, "treasury-key-first")); err != nil {
		t.Fatalf("first designation: %v", err)
	}

	_, err := treasury.Designate(context.Background(), secondOrganization, userID,
		domain.RoleOwner, ownership(message, signature, "treasury-key-second"))
	if !errors.Is(err, service.ErrProofRejected) {
		t.Fatalf("expected a consumed nonce to be refused, got %v", err)
	}
}

func TestNonceIssuedToAnotherUserIsRefused(t *testing.T) {
	handle := store(t)
	owner, _ := seedUser(t, handle)
	stranger, _ := seedUser(t, handle)
	organizationID := seedOrganization(t, handle, owner)

	treasury := treasuryService(handle)

	nonce, err := treasury.IssueNonce(context.Background(), stranger)
	if err != nil {
		t.Fatalf("issue nonce: %v", err)
	}

	key, address := walletKey(t)
	message, signature := signedProof(t, key, address, nonce.Value)

	_, err = treasury.Designate(context.Background(), organizationID, owner,
		domain.RoleOwner, ownership(message, signature, "treasury-key-stranger"))
	if !errors.Is(err, service.ErrProofRejected) {
		t.Fatalf("expected a nonce issued to another user to be refused, got %v", err)
	}
}

func TestSecondDesignationForTheSameOrganizationIsRefused(t *testing.T) {
	handle := store(t)
	userID, _ := seedUser(t, handle)
	organizationID := seedOrganization(t, handle, userID)

	treasury := treasuryService(handle)

	first, _ := treasury.IssueNonce(context.Background(), userID)
	keyOne, addressOne := walletKey(t)
	messageOne, signatureOne := signedProof(t, keyOne, addressOne, first.Value)

	if _, err := treasury.Designate(context.Background(), organizationID, userID,
		domain.RoleOwner, ownership(messageOne, signatureOne, "treasury-a")); err != nil {
		t.Fatalf("first designation: %v", err)
	}

	second, _ := treasury.IssueNonce(context.Background(), userID)
	keyTwo, addressTwo := walletKey(t)
	messageTwo, signatureTwo := signedProof(t, keyTwo, addressTwo, second.Value)

	_, err := treasury.Designate(context.Background(), organizationID, userID,
		domain.RoleOwner, ownership(messageTwo, signatureTwo, "treasury-b"))
	if !errors.Is(err, service.ErrTreasuryDesignated) {
		t.Fatalf("expected ErrTreasuryDesignated, got %v", err)
	}
}

func TestAddressCannotBeTheTreasuryOfTwoOrganizations(t *testing.T) {
	handle := store(t)

	firstUser, _ := seedUser(t, handle)
	firstOrganization := seedOrganization(t, handle, firstUser)

	secondUser, _ := seedUser(t, handle)
	secondOrganization := seedOrganization(t, handle, secondUser)

	treasury := treasuryService(handle)
	key, address := walletKey(t)

	nonceOne, _ := treasury.IssueNonce(context.Background(), firstUser)
	messageOne, signatureOne := signedProof(t, key, address, nonceOne.Value)

	if _, err := treasury.Designate(context.Background(), firstOrganization, firstUser,
		domain.RoleOwner, ownership(messageOne, signatureOne, "shared-a")); err != nil {
		t.Fatalf("first designation: %v", err)
	}

	nonceTwo, _ := treasury.IssueNonce(context.Background(), secondUser)
	messageTwo, signatureTwo := signedProof(t, key, address, nonceTwo.Value)

	_, err := treasury.Designate(context.Background(), secondOrganization, secondUser,
		domain.RoleOwner, ownership(messageTwo, signatureTwo, "shared-b"))
	if !errors.Is(err, service.ErrAddressTaken) {
		t.Fatalf("expected ErrAddressTaken, got %v", err)
	}
}

func lowerHex(address string) string {
	lowered := make([]byte, 0, len(address))
	for _, character := range []byte(address) {
		if character >= 'A' && character <= 'F' {
			character += 'a' - 'A'
		}
		lowered = append(lowered, character)
	}
	return string(lowered)
}
