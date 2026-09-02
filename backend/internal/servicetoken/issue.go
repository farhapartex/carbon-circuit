package servicetoken

import (
	"crypto/ed25519"
	"fmt"
	"time"

	jose "gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

type Signer struct {
	signer   jose.Signer
	lifetime time.Duration
}

func NewSigner(privateKey ed25519.PrivateKey, lifetime time.Duration) (*Signer, error) {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.EdDSA, Key: privateKey},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		return nil, fmt.Errorf("build service token signer: %w", err)
	}

	return &Signer{signer: signer, lifetime: lifetime}, nil
}

func (s *Signer) Issue(caller Caller) (string, error) {
	now := time.Now()

	token, err := jwt.Signed(s.signer).Claims(envelope{
		Caller:    caller,
		Issuer:    Issuer,
		Audience:  Audience,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.lifetime).Unix(),
	}).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("sign service token: %w", err)
	}

	return token, nil
}
