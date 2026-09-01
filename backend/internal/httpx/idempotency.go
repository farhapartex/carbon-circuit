package httpx

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	idempotencyHeader = "Idempotency-Key"
	minimumKeyLength  = 8
	maximumKeyLength  = 255
)

type idempotencyContextKey struct{}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyContextKey{}, key)
}

func IdempotencyKeyFrom(ctx context.Context) (string, bool) {
	key, present := ctx.Value(idempotencyContextKey{}).(string)
	return key, present
}

func mutating(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func usableKey(key string) bool {
	if len(key) < minimumKeyLength || len(key) > maximumKeyLength {
		return false
	}

	for _, character := range key {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}

	return true
}

func RequireIdempotencyKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mutating(c.Request.Method) {
			c.Next()
			return
		}

		key := strings.TrimSpace(c.GetHeader(idempotencyHeader))
		if key == "" {
			Fail(c, CodeIdempotencyKeyRequired)
			return
		}

		if !usableKey(key) {
			Fail(c, CodeValidation, FieldError{
				Field: idempotencyHeader,
				Code:  "UNSUPPORTED_VALUE",
			})
			return
		}

		c.Request = c.Request.WithContext(WithIdempotencyKey(c.Request.Context(), key))
		c.Next()
	}
}
