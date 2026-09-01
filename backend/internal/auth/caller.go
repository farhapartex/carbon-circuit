package auth

import (
	"context"
	"time"
)

type Caller struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	IssuedAt      time.Time
}

type callerContextKey struct{}

func WithCaller(ctx context.Context, caller Caller) context.Context {
	return context.WithValue(ctx, callerContextKey{}, caller)
}

func CallerFrom(ctx context.Context) (Caller, bool) {
	caller, present := ctx.Value(callerContextKey{}).(Caller)
	return caller, present
}
