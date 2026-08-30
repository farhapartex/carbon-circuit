package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/internal/logging"
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

	identity, err := upstream.DialIdentity(settings.IdentityAddress, settings.UpstreamCallTimeout)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := identity.Close(); closeErr != nil {
			logger.Error("closing identity client", slog.Any("error", closeErr))
		}
	}()

	logger.Info("identity upstream configured",
		slog.String("address", settings.IdentityAddress),
		slog.String("environment", settings.Environment),
	)

	router := handler.NewRouter(handler.RouterOptions{
		Identity:    identity,
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
