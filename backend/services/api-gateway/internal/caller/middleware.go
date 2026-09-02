package caller

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/internal/servicetoken"
)

type contextKey struct{}

func ContextFrom(ctx context.Context) (servicetoken.Caller, bool) {
	resolved, present := ctx.Value(contextKey{}).(servicetoken.Caller)
	return resolved, present
}

func withContext(ctx context.Context, resolved servicetoken.Caller) context.Context {
	return context.WithValue(ctx, contextKey{}, resolved)
}

func Stamp(
	resolver *Resolver,
	signer *servicetoken.Signer,
	logger *slog.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		verified, authenticated := auth.CallerFrom(c.Request.Context())
		if !authenticated {
			c.Next()
			return
		}

		resolved, err := resolver.Resolve(c.Request.Context(), verified)
		if err != nil {
			logger.Error("could not resolve caller context",
				slog.String("subject", verified.Subject),
				slog.String("request_id", httpx.CorrelationID(c)),
				slog.Any("error", err),
			)
			httpx.Fail(c, httpx.CodeDependencyUnavailable)
			return
		}

		token, err := signer.Issue(resolved)
		if err != nil {
			logger.Error("could not issue service token",
				slog.String("subject", verified.Subject),
				slog.Any("error", err),
			)
			httpx.Fail(c, httpx.CodeInternal)
			return
		}

		stamped := withContext(c.Request.Context(), resolved)
		c.Request = c.Request.WithContext(grpcx.WithServiceToken(stamped, token))
		c.Next()
	}
}
