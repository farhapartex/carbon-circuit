package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	provenancev1 "github.com/carboncircuit/backend/gen/carboncircuit/provenance/v1"
	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/provenance-service/internal/config"
	"github.com/carboncircuit/backend/services/provenance-service/internal/repository"
	"github.com/carboncircuit/backend/services/provenance-service/internal/rpc"
	"github.com/carboncircuit/backend/services/provenance-service/internal/service"
	"github.com/carboncircuit/backend/services/provenance-service/internal/upstream"
)

var revision = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("provenance-service failed to start", slog.Any("error", err))
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

	producer, err := kafka.NewProducer(kafka.ProducerOptions{
		Brokers:          settings.KafkaBrokers,
		AllowTopicCreate: settings.KafkaTopicCreate,
	})
	if err != nil {
		return err
	}
	defer producer.Close()

	if pingErr := producer.Ping(ctx); pingErr != nil {
		logger.Warn("kafka unreachable at startup, outbox will retry",
			slog.Any("error", pingErr))
	}

	publisherCtx, stopPublisher := context.WithCancel(ctx)
	defer stopPublisher()

	go outbox.NewPublisher(outbox.PublisherOptions{
		Database:  store,
		Dispatch:  producer,
		Logger:    logger,
		Interval:  settings.OutboxInterval,
		BatchSize: settings.OutboxBatchSize,
	}).Run(publisherCtx)

	publicKey, err := sharedconfig.Ed25519PublicKey(settings.ServiceTokenPublicKey)
	if err != nil {
		return err
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

	clientTransport, err := grpcx.ClientCredentials(settings.TLS, "identity-service")
	if err != nil {
		return err
	}

	identity, err := upstream.DialIdentity(
		settings.IdentityAddress, settings.CallTimeout, clientTransport,
	)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := identity.Close(); closeErr != nil {
			logger.Error("closing identity client", slog.Any("error", closeErr))
		}
	}()

	batches := service.NewBatchService(
		store,
		repository.NewBatchRepository(),
		repository.NewCheckpointRepository(),
		identity,
		logger,
	)

	provenanceServer := rpc.NewProvenanceServer(store, batches, logger, revision)

	logger.Info("provenance-service ready",
		slog.String("schema", settings.DatabaseSchema),
		slog.String("revision", revision),
	)

	return grpcx.Serve(ctx, grpcx.ServerOptions{
		Interceptors: []grpc.UnaryServerInterceptor{
			grpcx.RequireServiceToken(servicetoken.NewVerifier(publicKey), nil),
		},
		TransportCreds:  transport,
		Address:         settings.GRPCAddress,
		Logger:          logger,
		ShutdownTimeout: settings.ShutdownTimeout,
		ServiceName:     "carboncircuit.provenance.v1.ProvenanceService",
		HealthInterval:  settings.HealthInterval,
		ReportHealth: func(healthCtx context.Context) bool {
			pool, poolErr := store.DB()
			return poolErr == nil && pool.PingContext(healthCtx) == nil
		},
		Register: func(server *grpc.Server) {
			provenancev1.RegisterProvenanceServiceServer(server, provenanceServer)
		},
	})
}
