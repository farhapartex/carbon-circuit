package auth

import "context"

const ClaimNamespace = "https://carboncircuit.dev"

type profileClaims struct {
	Email         string `json:"https://carboncircuit.dev/email"`
	EmailVerified bool   `json:"https://carboncircuit.dev/email_verified"`
	Name          string `json:"https://carboncircuit.dev/name"`
}

func (profileClaims) Validate(context.Context) error { return nil }
