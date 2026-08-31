package httpx

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/ratelimit"
)

const EndpointClassKey = "endpoint_class"

func EndpointClass(class string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(EndpointClassKey, class)
		c.Next()
	}
}

func RateLimit(limiter *ratelimit.Limiter, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		endpointClass, _ := c.Get(EndpointClassKey)
		class, _ := endpointClass.(string)

		request := ratelimit.Request{
			CallerClass:   "public",
			CallerKey:     c.ClientIP(),
			EndpointClass: class,
			ClientIP:      c.ClientIP(),
		}

		if caller, verified := auth.CallerFrom(c.Request.Context()); verified {
			request.CallerClass = "user"
			request.CallerKey = caller.Subject
		}

		decision, err := limiter.Check(c.Request.Context(), request)
		if err != nil {
			logger.Warn("rate limiter unavailable, allowing request",
				slog.Any("error", err),
				slog.String("request_id", CorrelationID(c)),
			)
			c.Next()
			return
		}

		if decision.Limit > 0 {
			c.Header("X-RateLimit-Limit", strconv.Itoa(decision.Limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(max(decision.Remaining, 0)))
			c.Header("X-RateLimit-Reset", strconv.Itoa(int(time.Now().Add(decision.ResetAfter).Unix())))
		}

		if !decision.Allowed {
			retryAfter := int(decision.RetryAfter.Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			logger.Info("rate limited",
				slog.String("rule", decision.Rule),
				slog.String("client_ip", c.ClientIP()),
				slog.String("request_id", CorrelationID(c)),
			)
			Fail(c, CodeRateLimited)
			return
		}

		c.Next()
	}
}
