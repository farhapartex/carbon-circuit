package servicetoken

import (
	"crypto/ed25519"
	"errors"
	"time"

	jose "gopkg.in/go-jose/go-jose.v2"
	"gopkg.in/go-jose/go-jose.v2/jwt"
)

var ErrInvalidToken = errors.New("service token failed verification")

const allowedClockSkew = 5 * time.Second

type Verifier struct {
	publicKey ed25519.PublicKey
}

func NewVerifier(publicKey ed25519.PublicKey) *Verifier {
	return &Verifier{publicKey: publicKey}
}

func (v *Verifier) Verify(token string) (Caller, error) {
	parsed, err := jwt.ParseSigned(token)
	if err != nil {
		return Caller{}, ErrInvalidToken
	}

	if len(parsed.Headers) != 1 || parsed.Headers[0].Algorithm != string(jose.EdDSA) {
		return Caller{}, ErrInvalidToken
	}

	var claims envelope
	if err := parsed.Claims(v.publicKey, &claims); err != nil {
		return Caller{}, ErrInvalidToken
	}

	if !claims.valid(time.Now(), allowedClockSkew) {
		return Caller{}, ErrInvalidToken
	}

	return claims.Caller, nil
}
