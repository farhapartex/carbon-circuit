package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	provenancereadv1 "github.com/carboncircuit/backend/gen/carboncircuit/provenanceread/v1"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/config"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/repository"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/rpc"
	"github.com/carboncircuit/backend/services/provenance-read-service/internal/service"
)

var revision = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("provenance-read-service failed to start", slog.Any("error", err))
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

	publicStore := repository.NewPublicRepository()

	projection := service.NewProjection(
		store, publicStore, settings.ConsumerGroup, logger,
	)

	deadLetter, err := kafka.NewProducer(kafka.ProducerOptions{
		Brokers:          settings.KafkaBrokers,
		AllowTopicCreate: true,
	})
	if err != nil {
		return err
	}
	defer deadLetter.Close()

	consumer, err := kafka.NewConsumer(kafka.ConsumerOptions{
		Brokers:    settings.KafkaBrokers,
		Group:      settings.ConsumerGroup,
		Topics:     []string{events.TopicBatchCreated, events.TopicCheckpointLogged},
		Logger:     logger,
		Handle:     projection.Apply,
		DeadLetter: deadLetter,
	})
	if err != nil {
		return err
	}
	defer consumer.Close()

	consumerCtx, stopConsumer := context.WithCancel(ctx)
	defer stopConsumer()

	go consumer.Run(consumerCtx)

	logger.Info("consuming provenance events",
		slog.String("group", settings.ConsumerGroup),
	)

	readServer := rpc.NewProvenanceReadServer(
		store, service.NewPublicReader(store, publicStore, logger), logger, revision,
	)

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

	logger.Info("provenance-read-service ready",
		slog.String("schema", settings.DatabaseSchema),
		slog.String("revision", revision),
	)

	return grpcx.Serve(ctx, grpcx.ServerOptions{
		TransportCreds:  transport,
		Address:         settings.GRPCAddress,
		Logger:          logger,
		ShutdownTimeout: settings.ShutdownTimeout,
		ServiceName:     "carboncircuit.provenanceread.v1.ProvenanceReadService",
		HealthInterval:  settings.HealthInterval,
		ReportHealth: func(healthCtx context.Context) bool {
			pool, poolErr := store.DB()
			return poolErr == nil && pool.PingContext(healthCtx) == nil
		},
		Register: func(server *grpc.Server) {
			provenancereadv1.RegisterProvenanceReadServiceServer(server, readServer)
		},
	})
}
