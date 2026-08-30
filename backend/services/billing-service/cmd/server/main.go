package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"

	billingv1 "github.com/carboncircuit/backend/gen/carboncircuit/billing/v1"
	"github.com/carboncircuit/backend/internal/cache"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
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

	plans := service.NewPlanService(repository.NewPlanRepository(store), cacheClient)
	billingServer := rpc.NewBillingServer(plans, logger)

	logger.Info("billing-service ready",
		slog.String("schema", settings.DatabaseSchema),
		slog.String("revision", revision),
	)

	return grpcx.Serve(ctx, grpcx.ServerOptions{
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
