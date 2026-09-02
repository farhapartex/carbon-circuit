package servicetoken

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"

	jose "gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

func keys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()

	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keys: %v", err)
	}
	return public, private
}

func caller() Caller {
	return Caller{
		Subject:            "google-oauth2|1234",
		UserID:             "01a05cec-c301-7885-ae16-9fb26e3b1b04",
		OrganizationID:     "01a05840-0000-7000-8000-000000000001",
		Role:               "owner",
		PlanTier:           "growth",
		VerificationStatus: "verified",
		OrganizationState:  "active",
	}
}

func TestIssuedTokenVerifies(t *testing.T) {
	public, private := keys(t)

	signer, err := NewSigner(private, 30*time.Second)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}

	token, err := signer.Issue(caller())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	verified, err := NewVerifier(public).Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if verified != caller() {
		t.Fatalf("expected the claims to survive the round trip, got %+v", verified)
	}
}

func TestAnotherKeyCannotForgeAToken(t *testing.T) {
	public, _ := keys(t)
	_, attacker := keys(t)

	forger, err := NewSigner(attacker, 30*time.Second)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}

	forged, err := forger.Issue(caller())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if _, err := NewVerifier(public).Verify(forged); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected a foreign key to be rejected, got %v", err)
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	public, private := keys(t)

	signer, err := NewSigner(private, -time.Minute)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}

	expired, err := signer.Issue(caller())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	if _, err := NewVerifier(public).Verify(expired); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("expected an expired token to be rejected")
	}
}

func TestUnsignedAndHmacTokensAreRejected(t *testing.T) {
	public, private := keys(t)

	signer, err := NewSigner(private, 30*time.Second)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}

	genuine, err := signer.Issue(caller())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	hmacSigner, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: []byte("attacker-chosen-secret-value-32b")},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		t.Fatalf("hmac signer: %v", err)
	}

	hmacToken, err := jwt.Signed(hmacSigner).Claims(envelope{
		Caller:    caller(),
		Issuer:    Issuer,
		Audience:  Audience,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Minute).Unix(),
	}).CompactSerialize()
	if err != nil {
		t.Fatalf("hmac token: %v", err)
	}

	segments := strings.Split(genuine, ".")
	tampered := segments[0] + "." + segments[1] + "x." + segments[2]

	verifier := NewVerifier(public)

	for name, token := range map[string]string{
		"hmac signed":      hmacToken,
		"tampered payload": tampered,
		"empty":            "",
		"not a token":      "clearly-not-a-token",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := verifier.Verify(token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected rejection, got %v", err)
			}
		})
	}
}

func TestForeignIssuerOrAudienceIsRejected(t *testing.T) {
	public, private := keys(t)

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.EdDSA, Key: private},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		t.Fatalf("signer: %v", err)
	}

	verifier := NewVerifier(public)

	for name, claims := range map[string]envelope{
		"foreign issuer": {
			Caller: caller(), Issuer: "someone-else", Audience: Audience,
			IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Minute).Unix(),
		},
		"foreign audience": {
			Caller: caller(), Issuer: Issuer, Audience: "another-system",
			IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Minute).Unix(),
		},
		"no subject": {
			Caller: Caller{UserID: "x"}, Issuer: Issuer, Audience: Audience,
			IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Minute).Unix(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			token, signErr := jwt.Signed(signer).Claims(claims).CompactSerialize()
			if signErr != nil {
				t.Fatalf("sign: %v", signErr)
			}
			if _, err := verifier.Verify(token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected rejection, got %v", err)
			}
		})
	}
}
