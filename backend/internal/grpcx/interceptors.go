package grpcx

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/carboncircuit/backend/internal/logging"
)

const (
	CorrelationMetadataKey = "x-request-id"
	IdempotencyMetadataKey = "x-idempotency-key"
)

func CorrelationIDFromIncoming(ctx context.Context) string {
	incoming, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := incoming.Get(CorrelationMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, CorrelationMetadataKey, correlationID)
}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	if key == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, IdempotencyMetadataKey, key)
}

func IdempotencyKeyFromIncoming(ctx context.Context) string {
	incoming, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := incoming.Get(IdempotencyMetadataKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func CorrelateUnary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		correlationID := CorrelationIDFromIncoming(ctx)
		if correlationID == "" {
			correlationID = uuid.NewString()
		}
		return handler(logging.WithCorrelationID(ctx, correlationID), request)
	}
}

func RecoverUnary(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (response any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("handler panicked",
					slog.Any("panic", recovered),
					slog.String("method", info.FullMethod),
					slog.String("request_id", logging.CorrelationIDFrom(ctx)),
				)
				err = status.Error(codes.Internal, "INTERNAL_ERROR")
			}
		}()
		return handler(ctx, request)
	}
}

func LogUnary(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		started := time.Now()
		response, err := handler(ctx, request)

		logger.Info("rpc",
			slog.String("method", info.FullMethod),
			slog.String("code", status.Code(err).String()),
			slog.String("request_id", logging.CorrelationIDFrom(ctx)),
			slog.Duration("took", time.Since(started)),
		)

		return response, err
	}
}
