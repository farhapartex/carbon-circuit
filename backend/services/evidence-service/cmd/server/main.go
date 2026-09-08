package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	evidencev1 "github.com/carboncircuit/backend/gen/carboncircuit/evidence/v1"
	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/evidence-service/internal/config"
	"github.com/carboncircuit/backend/services/evidence-service/internal/inspect"
	"github.com/carboncircuit/backend/services/evidence-service/internal/repository"
	"github.com/carboncircuit/backend/services/evidence-service/internal/rpc"
	"github.com/carboncircuit/backend/services/evidence-service/internal/scan"
	"github.com/carboncircuit/backend/services/evidence-service/internal/service"
	"github.com/carboncircuit/backend/services/evidence-service/internal/storage"
)

var revision = "dev"

const pingMethod = "/carboncircuit.evidence.v1.EvidenceService/Ping"

func main() {
	if err := run(); err != nil {
		slog.Error("evidence-service failed to start", slog.Any("error", err))
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

	if settings.ScanningRequired() {
		return fmt.Errorf(
			"environment %q requires a malware scanner; set SCANNER_KIND=%s and SCANNER_ADDRESS",
			settings.Environment, scan.ClamAV)
	}

	scanner, err := scan.Open(settings.ScannerKind, settings.ScannerAddress, logger)
	if err != nil {
		return err
	}

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

	objects, err := storage.Open(ctx, storage.Settings{
		Endpoint:  settings.StorageEndpoint,
		AccessKey: settings.StorageAccessKey,
		SecretKey: settings.StorageSecretKey,
		Bucket:    settings.StorageBucket,
		UseTLS:    settings.StorageUseTLS,
		LinkTTL:   settings.DownloadLinkTTL,
	})
	if err != nil {
		return err
	}

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

	documents := service.NewDocumentService(
		store,
		repository.NewDocumentRepository(),
		objects,
		scanner,
		inspect.Limits{
			MaximumBytes: settings.MaximumByteSize,
			MaximumPages: settings.MaximumPages,
		},
		logger,
	)

	evidenceServer := rpc.NewEvidenceServer(
		store, documents, objects, scanner.Name(), settings.MaximumByteSize, logger, revision,
	)

	verifier := servicetoken.NewVerifier(publicKey)
	exempt := map[string]bool{pingMethod: true}

	logger.Info("evidence-service ready",
		slog.String("schema", settings.DatabaseSchema),
		slog.String("bucket", settings.StorageBucket),
		slog.String("scanner", scanner.Name()),
		slog.String("revision", revision),
	)

	return grpcx.Serve(ctx, grpcx.ServerOptions{
		Interceptors: []grpc.UnaryServerInterceptor{
			grpcx.RequireServiceToken(verifier, exempt),
		},
		StreamInterceptors: []grpc.StreamServerInterceptor{
			grpcx.RequireServiceTokenStream(verifier, exempt),
		},
		TransportCreds:  transport,
		Address:         settings.GRPCAddress,
		Logger:          logger,
		ShutdownTimeout: settings.ShutdownTimeout,
		ServiceName:     "carboncircuit.evidence.v1.EvidenceService",
		HealthInterval:  settings.HealthInterval,
		ReportHealth: func(healthCtx context.Context) bool {
			pool, poolErr := store.DB()
			return poolErr == nil && pool.PingContext(healthCtx) == nil && objects.Reachable(healthCtx)
		},
		Register: func(server *grpc.Server) {
			evidencev1.RegisterEvidenceServiceServer(server, evidenceServer)
		},
	})
}
