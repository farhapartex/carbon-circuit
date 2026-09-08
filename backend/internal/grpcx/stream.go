package grpcx

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/internal/servicetoken"
)

type reboundStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s reboundStream) Context() context.Context { return s.ctx }

func rebind(stream grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return reboundStream{ServerStream: stream, ctx: ctx}
}

func CorrelateStream() grpc.StreamServerInterceptor {
	return func(
		server any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		correlationID := CorrelationIDFromIncoming(stream.Context())
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		bound := logging.WithCorrelationID(stream.Context(), correlationID)

		return handler(server, rebind(stream, bound))
	}
}

func RecoverStream(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(
		server any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("stream handler panicked",
					slog.Any("panic", recovered),
					slog.String("method", info.FullMethod),
					slog.String("request_id", logging.CorrelationIDFrom(stream.Context())),
				)
				err = status.Error(codes.Internal, "INTERNAL_ERROR")
			}
		}()

		return handler(server, stream)
	}
}

func LogStream(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(
		server any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		started := time.Now()
		err := handler(server, stream)

		logger.Info("rpc stream",
			slog.String("method", info.FullMethod),
			slog.String("code", status.Code(err).String()),
			slog.String("request_id", logging.CorrelationIDFrom(stream.Context())),
			slog.Duration("took", time.Since(started)),
		)

		return err
	}
}

func RequireServiceTokenStream(
	verifier *servicetoken.Verifier,
	exempt map[string]bool,
) grpc.StreamServerInterceptor {
	return func(
		server any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if exempt[info.FullMethod] {
			return handler(server, stream)
		}

		token := serviceTokenFrom(stream.Context())
		if token == "" {
			return status.Error(codes.Unauthenticated, "a service token is required")
		}

		caller, err := verifier.Verify(token)
		if err != nil {
			return status.Error(codes.Unauthenticated, "service token rejected")
		}

		return handler(server, rebind(stream, withCaller(stream.Context(), caller)))
	}
}
