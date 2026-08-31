package httpx

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/auth"
)

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (auth.Caller, error)
}

type RevocationChecker interface {
	Revoked(ctx context.Context, caller auth.Caller) bool
}

const bearerPrefix = "Bearer "

func bearerToken(header string) (string, bool) {
	if len(header) <= len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(bearerPrefix):])
	return token, token != ""
}

func Authenticate(verifier TokenVerifier, denylist RevocationChecker, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, present := bearerToken(c.GetHeader("Authorization"))
		if !present {
			Fail(c, CodeUnauthenticated)
			return
		}

		caller, err := verifier.Verify(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, auth.ErrKeysUnavailable) {
				logger.Error("token signing keys unavailable",
					slog.String("request_id", CorrelationID(c)),
					slog.Any("error", err),
				)
				Fail(c, CodeDependencyUnavailable)
				return
			}

			logger.Info("token rejected",
				slog.String("request_id", CorrelationID(c)),
				slog.Any("error", err),
			)
			Fail(c, CodeUnauthenticated)
			return
		}

		if denylist.Revoked(c.Request.Context(), caller) {
			logger.Info("revoked token presented",
				slog.String("request_id", CorrelationID(c)),
				slog.String("subject", caller.Subject),
			)
			Fail(c, CodeTokenRevoked)
			return
		}

		c.Request = c.Request.WithContext(auth.WithCaller(c.Request.Context(), caller))
		c.Next()
	}
}
