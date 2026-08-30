package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const correlationIDKey contextKey = "correlation_id"

func New(serviceName, level string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	})
	return slog.New(handler).With(slog.String("service", serviceName))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

func CorrelationIDFrom(ctx context.Context) string {
	value, _ := ctx.Value(correlationIDKey).(string)
	return value
}

func FromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	correlationID := CorrelationIDFrom(ctx)
	if correlationID == "" {
		return base
	}
	return base.With(slog.String("request_id", correlationID))
}
