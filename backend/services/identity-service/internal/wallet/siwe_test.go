package wallet

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spruceid/siwe-go"
)

const (
	testDomain = "localhost:3000"
	testChain  = 31337
	testNonce  = "abcdef0123456789"
)

func signer(t *testing.T) (*ecdsa.PrivateKey, string) {
	t.Helper()

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	return key, crypto.PubkeyToAddress(key.PublicKey).Hex()
}

func message(t *testing.T, address, domain, nonce string, chainID int) *siwe.Message {
	t.Helper()

	issued := time.Now().UTC().Format(time.RFC3339)
	statement := "Designate this address as your CarbonCircuit Treasury Address."

	built, err := siwe.InitMessage(domain, address, "http://"+domain, nonce, map[string]any{
		"chainId":   chainID,
		"statement": statement,
		"issuedAt":  issued,
	})
	if err != nil {
		t.Fatalf("build message: %v", err)
	}

	return built
}

func sign(t *testing.T, key *ecdsa.PrivateKey, raw string) string {
	t.Helper()

	hash := accounts.TextHash([]byte(raw))
	signature, err := crypto.Sign(hash, key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	signature[64] += 27
	return fmt.Sprintf("0x%x", signature)
}

func expectation() Expectation {
	return Expectation{Domain: testDomain, ChainID: testChain}
}

func TestGenuineSignatureYieldsTheSigningAddress(t *testing.T) {
	key, address := signer(t)
	built := message(t, address, testDomain, testNonce, testChain)
	raw := built.String()

	proof, err := Verify(expectation(), raw, sign(t, key, raw))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if proof.Address != lower(address) {
		t.Fatalf("expected %s, got %s", lower(address), proof.Address)
	}
	if proof.Nonce != testNonce {
		t.Fatalf("expected the nonce to be carried, got %s", proof.Nonce)
	}
}

func TestAnotherKeysSignatureIsRejected(t *testing.T) {
	_, address := signer(t)
	attacker, _ := signer(t)

	built := message(t, address, testDomain, testNonce, testChain)
	raw := built.String()

	if _, err := Verify(expectation(), raw, sign(t, attacker, raw)); !errors.Is(err, ErrSignatureInvalid) {
		t.Fatalf("expected ErrSignatureInvalid, got %v", err)
	}
}

func TestForeignDomainIsRejected(t *testing.T) {
	key, address := signer(t)
	built := message(t, address, "evil.example", testNonce, testChain)
	raw := built.String()

	if _, err := Verify(expectation(), raw, sign(t, key, raw)); !errors.Is(err, ErrDomainMismatch) {
		t.Fatalf("expected ErrDomainMismatch, got %v", err)
	}
}

func TestForeignChainIsRejected(t *testing.T) {
	key, address := signer(t)
	built := message(t, address, testDomain, testNonce, 8453)
	raw := built.String()

	if _, err := Verify(expectation(), raw, sign(t, key, raw)); !errors.Is(err, ErrChainMismatch) {
		t.Fatalf("expected ErrChainMismatch, got %v", err)
	}
}

func TestTamperedMessageIsRejected(t *testing.T) {
	key, address := signer(t)
	built := message(t, address, testDomain, testNonce, testChain)
	raw := built.String()
	signature := sign(t, key, raw)

	tampered := message(t, address, testDomain, "9999999999999999", testChain).String()

	if _, err := Verify(expectation(), tampered, signature); err == nil {
		t.Fatal("expected a signature over a different message to be rejected")
	}
}

func TestUnreadableMessageIsRejected(t *testing.T) {
	for name, raw := range map[string]string{
		"empty":    "",
		"not siwe": "hello world",
		"json":     `{"domain":"localhost:3000"}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Verify(expectation(), raw, "0xdeadbeef"); !errors.Is(err, ErrMessageUnreadable) {
				t.Fatalf("expected ErrMessageUnreadable, got %v", err)
			}
		})
	}
}

func lower(address string) string {
	return Proof{Address: address}.lowered()
}
