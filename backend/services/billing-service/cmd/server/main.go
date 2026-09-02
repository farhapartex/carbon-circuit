package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/cache"
	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/billing-service/internal/config"
	"github.com/carboncircuit/backend/services/billing-service/internal/repository"
	"github.com/carboncircuit/backend/services/billing-service/internal/rpc"
	"github.com/carboncircuit/backend/services/billing-service/internal/service"
)

var revision = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("billing-service failed to start", slog.Any("error", err))
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

	store, err := database.Open(ctx, database.Options{
		DSN:             settings.DatabaseDSN,
		Schema:          settings.DatabaseSchema,
		MaxOpenConns:    settings.MaxOpenConns,
		MaxIdleConns:    settings.MaxIdleConns,
		ConnMaxLifetime: settings.ConnMaxLifetime,
		ConnMaxIdleTime: settings.ConnMaxIdleTime,
		AcquireTimeout:  settings.AcquireTimeout,
	})
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := database.Close(store); closeErr != nil {
			logger.Error("closing database", slog.Any("error", closeErr))
		}
	}()

	cacheClient := cache.New(settings.RedisAddress, settings.RedisPassword, settings.RedisDatabase, logger)
	defer func() {
		if closeErr := cacheClient.Close(); closeErr != nil {
			logger.Error("closing cache", slog.Any("error", closeErr))
		}
	}()

	if pingErr := cacheClient.Ping(ctx); pingErr != nil {
		logger.Warn("redis unreachable at startup, serving from postgres", slog.Any("error", pingErr))
	}

	subscriptionStore := repository.NewSubscriptionRepository(store)

	plans := service.NewPlanService(repository.NewPlanRepository(store), cacheClient)
	billingServer := rpc.NewBillingServer(
		plans,
		service.NewSubscriptionService(subscriptionStore),
		service.NewSubscriptionCreator(store, subscriptionStore, logger),
		logger,
	)

	logger.Info("billing-service ready",
		slog.String("schema", settings.DatabaseSchema),
		slog.String("revision", revision),
	)

	publicKey, err := sharedconfig.Ed25519PublicKey(settings.ServiceTokenPublicKey)
	if err != nil {
		return err
	}

	exempt := map[string]bool{
		"/carboncircuit.billing.v1.BillingService/ListPlans": true,
	}

	var transport credentials.TransportCredentials
	if settings.TLS.Configured() {
		transport, err = grpcx.ServerCredentials(settings.TLS)
		if err != nil {
			return err
		}
		logger.Info("serving grpc over mutual tls")
	} else {
		logger.Warn("grpc is serving without mutual tls")
	}

	return grpcx.Serve(ctx, grpcx.ServerOptions{
		Interceptors: []grpc.UnaryServerInterceptor{
			grpcx.RequireServiceToken(servicetoken.NewVerifier(publicKey), exempt),
		},
		TransportCreds:  transport,
		Address:         settings.GRPCAddress,
		Logger:          logger,
		ShutdownTimeout: settings.ShutdownTimeout,
		ServiceName:     "carboncircuit.billing.v1.BillingService",
		HealthInterval:  settings.HealthInterval,
		ReportHealth: func(healthCtx context.Context) bool {
			pool, poolErr := store.DB()
			return poolErr == nil && pool.PingContext(healthCtx) == nil
		},
		Register: func(server *grpc.Server) {
			billingv1.RegisterBillingServiceServer(server, billingServer)
		},
	})
}
