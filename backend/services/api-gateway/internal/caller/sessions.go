package caller

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/cache"
	"github.com/carboncircuit/backend/internal/grpcx"
)

const (
	sightingKeyPrefix  = "session:seen:"
	untrackedKeyPrefix = "session:untracked:"
)

type SessionRecorder interface {
	RecordSession(ctx context.Context, userAgent, ipAddress string) error
}

func sightingKey(subject, sessionID string) string {
	return sightingKeyPrefix + subject + ":" + sessionID
}

func RecordSessions(
	recorder SessionRecorder,
	client *cache.Client,
	interval time.Duration,
	logger *slog.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		verified, authenticated := auth.CallerFrom(c.Request.Context())
		if !authenticated {
			return
		}

		if verified.SessionID == "" {
			reportUntracked(c, client, interval, logger, verified.Subject)
			return
		}

		key := sightingKey(verified.Subject, verified.SessionID)

		seenValue, seen, err := client.GetString(c.Request.Context(), key)
		if err != nil {
			logger.Warn("could not read the session sighting throttle",
				slog.Any("error", err),
			)
			return
		}
		if seen {
			_ = seenValue
			return
		}

		if err := client.SetString(c.Request.Context(), key, "1", interval); err != nil {
			logger.Warn("could not throttle session recording", slog.Any("error", err))
			return
		}

		token := grpcx.ServiceTokenFrom(c.Request.Context())
		userAgent := c.Request.UserAgent()
		ipAddress := c.ClientIP()

		go func() {
			recordCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := recorder.RecordSession(
				grpcx.WithServiceToken(recordCtx, token), userAgent, ipAddress,
			); err != nil {
				logger.Warn("could not record session sighting", slog.Any("error", err))
			}
		}()
	}
}

func reportUntracked(
	c *gin.Context,
	client *cache.Client,
	interval time.Duration,
	logger *slog.Logger,
	subject string,
) {
	key := untrackedKeyPrefix + subject

	if _, reported, err := client.GetString(c.Request.Context(), key); err != nil || reported {
		return
	}

	if err := client.SetString(c.Request.Context(), key, "1", interval); err != nil {
		return
	}

	logger.Warn("token carries no session id, this device cannot be tracked",
		slog.String("subject", subject),
	)
}
