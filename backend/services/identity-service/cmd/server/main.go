package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"

	identityv1 "github.com/carboncircuit/backend/gen/carboncircuit/identity/v1"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/services/identity-service/internal/config"
	"github.com/carboncircuit/backend/services/identity-service/internal/repository"
	"github.com/carboncircuit/backend/services/identity-service/internal/rpc"
	"github.com/carboncircuit/backend/services/identity-service/internal/service"
)

var revision = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("identity-service failed to start", slog.Any("error", err))
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

	logger.Info("connected to postgres",
		slog.String("schema", settings.DatabaseSchema),
		slog.String("environment", settings.Environment),
	)

	sessions := service.NewSessionService(
		repository.NewUserRepository(store),
		repository.NewMembershipRepository(store),
		logger,
	)

	identityServer := rpc.NewIdentityServer(store, sessions, logger, revision)

	return grpcx.Serve(ctx, grpcx.ServerOptions{
		Address:         settings.GRPCAddress,
		Logger:          logger,
		ShutdownTimeout: settings.ShutdownTimeout,
		ServiceName:     "carboncircuit.identity.v1.IdentityService",
		HealthInterval:  settings.HealthInterval,
		ReportHealth:    identityServer.DatabaseReachable,
		Register: func(server *grpc.Server) {
			identityv1.RegisterIdentityServiceServer(server, identityServer)
		},
	})
}
