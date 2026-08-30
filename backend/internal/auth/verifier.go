package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

var (
	ErrInvalidToken    = errors.New("token failed verification")
	ErrKeysUnavailable = errors.New("signing keys unavailable")
)

const allowedClockSkew = 30 * time.Second

type Verifier struct {
	keys      func(context.Context) (any, error)
	validator *validator.Validator
}

func NewVerifier(domain, audience string, keyCacheTTL time.Duration) (*Verifier, error) {
	issuer, err := url.Parse("https://" + domain + "/")
	if err != nil {
		return nil, fmt.Errorf("parse auth0 issuer: %w", err)
	}

	provider := jwks.NewCachingProvider(issuer, keyCacheTTL)

	tokenValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuer.String(),
		[]string{audience},
		validator.WithAllowedClockSkew(allowedClockSkew),
	)
	if err != nil {
		return nil, fmt.Errorf("build token validator: %w", err)
	}

	return &Verifier{keys: provider.KeyFunc, validator: tokenValidator}, nil
}

func (v *Verifier) Verify(ctx context.Context, token string) (Caller, error) {
	if _, err := v.keys(ctx); err != nil {
		return Caller{}, fmt.Errorf("%w: %v", ErrKeysUnavailable, err)
	}

	verified, err := v.validator.ValidateToken(ctx, token)
	if err != nil {
		return Caller{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, recognised := verified.(*validator.ValidatedClaims)
	if !recognised {
		return Caller{}, fmt.Errorf("%w: unrecognised claim shape", ErrInvalidToken)
	}

	if claims.RegisteredClaims.Subject == "" {
		return Caller{}, fmt.Errorf("%w: token carries no subject", ErrInvalidToken)
	}

	if claims.RegisteredClaims.IssuedAt == 0 {
		return Caller{}, fmt.Errorf("%w: token carries no issued-at", ErrInvalidToken)
	}

	return Caller{
		Subject:  claims.RegisteredClaims.Subject,
		IssuedAt: time.Unix(claims.RegisteredClaims.IssuedAt, 0),
	}, nil
}
