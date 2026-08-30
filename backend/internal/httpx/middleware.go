package httpx

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/carboncircuit/backend/internal/logging"
)

const (
	correlationHeader  = "X-Request-Id"
	correlationContext = "correlation_id"
)

func CorrelationID(c *gin.Context) string {
	value, _ := c.Get(correlationContext)
	id, _ := value.(string)
	return id
}

func Correlate() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader(correlationHeader)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		c.Set(correlationContext, correlationID)
		c.Header(correlationHeader, correlationID)
		c.Request = c.Request.WithContext(
			logging.WithCorrelationID(c.Request.Context(), correlationID),
		)
		c.Next()
	}
}

func RecoverPanics(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("handler panicked",
					slog.Any("panic", recovered),
					slog.String("request_id", CorrelationID(c)),
					slog.String("path", c.FullPath()),
				)
				Fail(c, CodeInternal)
			}
		}()
		c.Next()
	}
}

func LogRequests(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		logger.Info("request",
			slog.String("request_id", CorrelationID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("took", time.Since(started)),
		)
	}
}
