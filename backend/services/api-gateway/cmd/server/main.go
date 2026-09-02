package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/carboncircuit/backend/internal/auth"
	"github.com/carboncircuit/backend/internal/cache"
	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/internal/ratelimit"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/api-gateway/internal/caller"
	"github.com/carboncircuit/backend/services/api-gateway/internal/config"
	"github.com/carboncircuit/backend/services/api-gateway/internal/handler"
	"github.com/carboncircuit/backend/services/api-gateway/internal/upstream"
)

var revision = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("api-gateway failed to start", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	settings, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(config.ServiceName, settings.LogLevel)
	ctx := context.Background()

	identityCreds, err := upstreamCredentials(settings, "identity-service", logger)
	if err != nil {
		return err
	}

	identity, err := upstream.DialIdentity(settings.IdentityAddress, settings.UpstreamCallTimeout, identityCreds)
	if err != nil {
		return err
	}
	defer closeUpstream(logger, "identity", identity.Close)

	billingCreds, err := upstreamCredentials(settings, "billing-service", logger)
	if err != nil {
		return err
	}

	billing, err := upstream.DialBilling(settings.BillingAddress, settings.UpstreamCallTimeout, billingCreds)
	if err != nil {
		return err
	}
	defer closeUpstream(logger, "billing", billing.Close)

	cacheClient := cache.New(settings.RedisAddress, settings.RedisPassword, settings.RedisDatabase, logger)
	defer closeUpstream(logger, "redis", cacheClient.Close)

	if pingErr := cacheClient.Ping(ctx); pingErr != nil {
		logger.Warn("redis unreachable at startup, rate limiting will fail open",
			slog.Any("error", pingErr))
	}

	limiter, err := ratelimit.New(cacheClient.Redis(), "ratelimit", publicRules(settings))
	if err != nil {
		return err
	}

	verifier, err := auth.NewVerifier(settings.Auth0Domain, settings.Auth0Audience, settings.KeyCacheTTL)
	if err != nil {
		return err
	}

	privateKey, err := sharedconfig.Ed25519PrivateKey(settings.ServiceTokenSeed)
	if err != nil {
		return err
	}

	signer, err := servicetoken.NewSigner(privateKey, settings.TokenLifetime)
	if err != nil {
		return err
	}

	resolver := caller.NewResolver(cacheClient, identity, settings.CallerContextTTL)

	denylist := auth.NewDenylist(cacheClient, logger, settings.RevocationWindow)

	logger.Info("api-gateway ready",
		slog.String("identity", settings.IdentityAddress),
		slog.String("billing", settings.BillingAddress),
		slog.Int("public_read_per_minute", settings.PublicReadPerMinute),
	)

	router := handler.NewRouter(handler.RouterOptions{
		Identity:    identity,
		Billing:     billing,
		Limiter:     limiter,
		Verifier:    verifier,
		Denylist:    denylist,
		Resolver:    resolver,
		Signer:      signer,
		Logger:      logger,
		Environment: settings.Environment,
		Revision:    revision,
	})

	return httpx.Serve(ctx, httpx.ServerOptions{
		Address:         settings.HTTPAddress,
		Handler:         router,
		Logger:          logger,
		ReadTimeout:     settings.ReadTimeout,
		WriteTimeout:    settings.WriteTimeout,
		ShutdownTimeout: settings.ShutdownTimeout,
	})
}

func publicRules(settings config.Config) []ratelimit.Rule {
	return []ratelimit.Rule{
		{
			Name:      "unauthenticated_public",
			PerMinute: settings.PublicReadPerMinute,
			Burst:     settings.PublicReadBurst,
			KeyFunc:   func(request ratelimit.Request) string { return "public:ip:" + request.ClientIP },
			AppliesTo: func(request ratelimit.Request) bool { return request.CallerClass == "public" },
		},
		{
			Name:      "portal_user",
			PerMinute: settings.PortalUserPerMinute,
			Burst:     settings.PortalUserBurst,
			KeyFunc:   func(request ratelimit.Request) string { return "portal:user:" + request.CallerKey },
			AppliesTo: func(request ratelimit.Request) bool { return request.CallerClass == "user" },
		},
	}
}

func closeUpstream(logger *slog.Logger, name string, close func() error) {
	if err := close(); err != nil {
		logger.Error("closing "+name, slog.Any("error", err))
	}
}

func upstreamCredentials(
	settings config.Config,
	service string,
	logger *slog.Logger,
) (credentials.TransportCredentials, error) {
	if !settings.TLS.Configured() {
		logger.Warn("dialling without mutual tls", slog.String("service", service))
		return insecure.NewCredentials(), nil
	}

	return grpcx.ClientCredentials(settings.TLS, service)
}
