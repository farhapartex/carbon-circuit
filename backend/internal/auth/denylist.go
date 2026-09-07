package auth

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/carboncircuit/backend/internal/cache"
)

const (
	revocationKeyPrefix = "auth:revoked:subject:"
	sessionKeyPrefix    = "auth:revoked:session:"
)

type Denylist struct {
	cache  *cache.Client
	logger *slog.Logger
	window time.Duration
}

func NewDenylist(client *cache.Client, logger *slog.Logger, window time.Duration) *Denylist {
	return &Denylist{cache: client, logger: logger, window: window}
}

func revocationKey(subject string) string {
	return revocationKeyPrefix + subject
}

func sessionKey(subject, sessionID string) string {
	return sessionKeyPrefix + subject + ":" + sessionID
}

func (d *Denylist) Revoke(ctx context.Context, subject string) error {
	revokedAt := strconv.FormatInt(time.Now().Unix(), 10)
	return d.cache.SetString(ctx, revocationKey(subject), revokedAt, d.window)
}

func (d *Denylist) RevokeSession(
	ctx context.Context,
	subject, sessionID string,
) error {
	if sessionID == "" {
		return nil
	}
	revokedAt := strconv.FormatInt(time.Now().Unix(), 10)
	return d.cache.SetString(ctx, sessionKey(subject, sessionID), revokedAt, d.window)
}

func (d *Denylist) Revoked(ctx context.Context, caller Caller) bool {
	if d.revokedAt(ctx, sessionKey(caller.Subject, caller.SessionID), caller) {
		return true
	}
	return d.revokedAt(ctx, revocationKey(caller.Subject), caller)
}

func (d *Denylist) revokedAt(ctx context.Context, key string, caller Caller) bool {
	if caller.SessionID == "" && key != revocationKey(caller.Subject) {
		return false
	}

	raw, found, err := d.cache.GetString(ctx, key)
	if err != nil {
		d.logger.Error("revocation denylist unreachable, admitting caller",
			slog.String("subject", caller.Subject),
			slog.Any("error", err),
		)
		return false
	}

	if !found {
		return false
	}

	revokedAt, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		d.logger.Error("revocation entry unreadable, admitting caller",
			slog.String("subject", caller.Subject),
			slog.Any("error", err),
		)
		return false
	}

	return !caller.IssuedAt.After(time.Unix(revokedAt, 0))
}
