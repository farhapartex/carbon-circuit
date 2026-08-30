package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/httpx"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/services/identity-service/internal/config"
	"github.com/carboncircuit/backend/services/identity-service/internal/handler"
)

var revision = "dev"

func main() {
	healthCheck := flag.Bool("health", false, "probe the local health endpoint and exit")
	flag.Parse()

	if *healthCheck {
		os.Exit(probeHealth())
	}

	if err := run(); err != nil {
		slog.Error("identity-service failed to start", slog.Any("error", err))
		os.Exit(1)
	}
}

func probeHealth() int {
	client := &http.Client{Timeout: 3 * time.Second}
	address := os.Getenv("HTTP_ADDRESS")
	if address == "" {
		address = ":8081"
	}

	response, err := client.Get("http://127.0.0.1" + address + "/readyz")
	if err != nil {
		return 1
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return 1
	}
	return 0
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

	router := handler.NewRouter(handler.RouterOptions{
		Database:    store,
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
