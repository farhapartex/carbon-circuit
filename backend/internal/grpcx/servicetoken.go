package grpcx

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/carboncircuit/backend/internal/servicetoken"
)

const ServiceTokenMetadataKey = "x-service-token"

type callerContextKey struct{}

func WithServiceToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, ServiceTokenMetadataKey, token)
}

func CallerFrom(ctx context.Context) (servicetoken.Caller, bool) {
	caller, present := ctx.Value(callerContextKey{}).(servicetoken.Caller)
	return caller, present
}

func withCaller(ctx context.Context, caller servicetoken.Caller) context.Context {
	return context.WithValue(ctx, callerContextKey{}, caller)
}

func serviceTokenFrom(ctx context.Context) string {
	incoming, present := metadata.FromIncomingContext(ctx)
	if !present {
		return ""
	}

	values := incoming.Get(ServiceTokenMetadataKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func RequireServiceToken(
	verifier *servicetoken.Verifier,
	exempt map[string]bool,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if exempt[info.FullMethod] {
			return handler(ctx, request)
		}

		token := serviceTokenFrom(ctx)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "a service token is required")
		}

		caller, err := verifier.Verify(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "service token rejected")
		}

		return handler(withCaller(ctx, caller), request)
	}
}

func ForwardServiceToken(ctx context.Context) context.Context {
	return WithServiceToken(ctx, serviceTokenFrom(ctx))
}
