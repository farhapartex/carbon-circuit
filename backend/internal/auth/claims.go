package auth

import "context"

const ClaimNamespace = "https://carboncircuit.dev"

type profileClaims struct {
	Email         string `json:"https://carboncircuit.dev/email"`
	EmailVerified bool   `json:"https://carboncircuit.dev/email_verified"`
	Name          string `json:"https://carboncircuit.dev/name"`
	SessionID     string `json:"https://carboncircuit.dev/sid"`
}

func (profileClaims) Validate(context.Context) error { return nil }
