package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	jose "gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

const (
	testAudience = "https://api.carboncircuit.test"
	testKeyID    = "test-signing-key"
	testSubject  = "auth0|abc123"
)

type issuer struct {
	url        string
	privateKey *rsa.PrivateKey
	close      func()
}

func newIssuer(t *testing.T) *issuer {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{
			"issuer":   server.URL + "/",
			"jwks_uri": server.URL + "/.well-known/jwks.json",
		})
	})

	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
			Key:       privateKey.Public(),
			KeyID:     testKeyID,
			Algorithm: string(jose.RS256),
			Use:       "sig",
		}}})
	})

	return &issuer{url: server.URL + "/", privateKey: privateKey, close: server.Close}
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func (i *issuer) verifier(t *testing.T) *Verifier {
	t.Helper()

	parsed, err := url.Parse(i.url)
	if err != nil {
		t.Fatalf("parse issuer url: %v", err)
	}

	verifier, err := newVerifier(parsed, testAudience, time.Minute)
	if err != nil {
		t.Fatalf("build verifier: %v", err)
	}
	return verifier
}

type claimOverrides struct {
	issuer   string
	audience string
	subject  string
	issuedAt *time.Time
	expiry   time.Time
}

func (i *issuer) sign(t *testing.T, overrides claimOverrides) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: i.privateKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", testKeyID),
	)
	if err != nil {
		t.Fatalf("build signer: %v", err)
	}

	claims := i.claimsFrom(overrides)

	token, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func (i *issuer) claimsFrom(overrides claimOverrides) jwt.Claims {
	now := time.Now()

	claims := jwt.Claims{
		Issuer:   i.url,
		Audience: jwt.Audience{testAudience},
		Subject:  testSubject,
		IssuedAt: jwt.NewNumericDate(now),
		Expiry:   jwt.NewNumericDate(now.Add(10 * time.Minute)),
	}

	if overrides.issuer != "" {
		claims.Issuer = overrides.issuer
	}
	if overrides.audience != "" {
		claims.Audience = jwt.Audience{overrides.audience}
	}
	if overrides.subject != "" {
		claims.Subject = overrides.subject
	}
	if overrides.subject == "-" {
		claims.Subject = ""
	}
	if overrides.issuedAt != nil {
		claims.IssuedAt = jwt.NewNumericDate(*overrides.issuedAt)
	}
	if !overrides.expiry.IsZero() {
		claims.Expiry = jwt.NewNumericDate(overrides.expiry)
	}
	return claims
}

func TestVerifyAcceptsGenuineToken(t *testing.T) {
	source := newIssuer(t)
	defer source.close()

	caller, err := source.verifier(t).Verify(context.Background(), source.sign(t, claimOverrides{}))
	if err != nil {
		t.Fatalf("expected token to verify, got %v", err)
	}

	if caller.Subject != testSubject {
		t.Fatalf("expected subject %q, got %q", testSubject, caller.Subject)
	}

	if caller.IssuedAt.IsZero() {
		t.Fatal("expected issued-at to be populated")
	}
}

func TestVerifyExtractsNamespacedProfileClaims(t *testing.T) {
	source := newIssuer(t)
	defer source.close()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: source.privateKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", testKeyID),
	)
	if err != nil {
		t.Fatalf("build signer: %v", err)
	}

	profile := map[string]any{
		ClaimNamespace + "/email":          "nazmul@example.test",
		ClaimNamespace + "/email_verified": true,
		ClaimNamespace + "/name":           "Nazmul",
	}

	token, err := jwt.Signed(signer).
		Claims(source.claimsFrom(claimOverrides{})).
		Claims(profile).
		CompactSerialize()
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	caller, err := source.verifier(t).Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if caller.Email != "nazmul@example.test" {
		t.Fatalf("expected email claim, got %q", caller.Email)
	}
	if !caller.EmailVerified {
		t.Fatal("expected email_verified claim to be carried through")
	}
	if caller.Name != "Nazmul" {
		t.Fatalf("expected name claim, got %q", caller.Name)
	}
}

func TestVerifyToleratesAbsentProfileClaims(t *testing.T) {
	source := newIssuer(t)
	defer source.close()

	caller, err := source.verifier(t).Verify(context.Background(), source.sign(t, claimOverrides{}))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if caller.Email != "" || caller.EmailVerified || caller.Name != "" {
		t.Fatal("expected empty profile when the action has not been deployed")
	}
}

func TestVerifyRejectsTamperedAndForgedTokens(t *testing.T) {
	source := newIssuer(t)
	defer source.close()

	verifier := source.verifier(t)
	genuine := source.sign(t, claimOverrides{})

	foreign := newIssuer(t)
	defer foreign.close()

	past := time.Now().Add(-2 * time.Hour)

	cases := map[string]string{
		"empty":            "",
		"not a jwt":        "clearly-not-a-token",
		"tampered payload": tamperPayload(genuine),
		"unsigned":         unsignedVariant(t, source),
		"foreign signer":   foreign.sign(t, claimOverrides{issuer: source.url}),
		"wrong audience":   source.sign(t, claimOverrides{audience: "https://api.someone-else.test"}),
		"wrong issuer":     source.sign(t, claimOverrides{issuer: "https://evil.example.com/"}),
		"expired":          source.sign(t, claimOverrides{expiry: past}),
		"no subject":       source.sign(t, claimOverrides{subject: "-"}),
		"no issued at":     source.sign(t, claimOverrides{issuedAt: &time.Time{}}),
	}

	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := verifier.Verify(context.Background(), token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("expected ErrInvalidToken, got %v", err)
			}
		})
	}
}

func tamperPayload(token string) string {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return token
	}
	return segments[0] + "." + segments[1] + "x." + segments[2]
}

func unsignedVariant(t *testing.T, source *issuer) string {
	t.Helper()

	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.HS256, Key: []byte("attacker-chosen-secret-value-32b")},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", testKeyID),
	)
	if err != nil {
		t.Fatalf("build hmac signer: %v", err)
	}

	token, err := jwt.Signed(signer).Claims(source.claimsFrom(claimOverrides{})).CompactSerialize()
	if err != nil {
		t.Fatalf("sign hmac token: %v", err)
	}
	return token
}

func TestVerifyReportsUnreachableKeysDistinctly(t *testing.T) {
	source := newIssuer(t)
	verifier := source.verifier(t)
	token := source.sign(t, claimOverrides{})
	source.close()

	_, err := verifier.Verify(context.Background(), token)
	if !errors.Is(err, ErrKeysUnavailable) {
		t.Fatalf("expected ErrKeysUnavailable, got %v", err)
	}
}
