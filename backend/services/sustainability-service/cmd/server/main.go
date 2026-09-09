package main

import (
	"context"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sustainabilityv1 "github.com/carboncircuit/backend/gen/carboncircuit/sustainability/v1"
	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/database"
	"github.com/carboncircuit/backend/internal/events"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/internal/kafka"
	"github.com/carboncircuit/backend/internal/logging"
	"github.com/carboncircuit/backend/internal/outbox"
	"github.com/carboncircuit/backend/internal/servicetoken"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/config"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/repository"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/rpc"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/service"
	"github.com/carboncircuit/backend/services/sustainability-service/internal/upstream"
)

var revision = "dev"

const pingMethod = "/carboncircuit.sustainability.v1.SustainabilityService/Ping"

func main() {
	if err := run(); err != nil {
		slog.Error("sustainability-service failed to start", slog.Any("error", err))
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

	identityCreds, err := grpcx.ClientCredentials(settings.TLS, "identity-service")
	if err != nil {
		return err
	}

	identity, err := upstream.DialIdentity(settings.IdentityAddress, settings.CallTimeout, identityCreds)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := identity.Close(); closeErr != nil {
			logger.Error("closing identity client", slog.Any("error", closeErr))
		}
	}()

	evidenceCreds, err := grpcx.ClientCredentials(settings.TLS, "evidence-service")
	if err != nil {
		return err
	}

	evidence, err := upstream.DialEvidence(settings.EvidenceAddress, settings.CallTimeout, evidenceCreds)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := evidence.Close(); closeErr != nil {
			logger.Error("closing evidence client", slog.Any("error", closeErr))
		}
	}()

	claims := service.NewClaimService(
		store,
		repository.NewClaimRepository(),
		repository.NewReferenceRepository(),
		identity,
		evidence,
		logger,
	)

	reviews := service.NewReviewPipeline(
		store,
		repository.NewClaimRepository(),
		settings.ConsumerGroup,
		settings.AIReviewStub,
		logger,
	)

	consumer, err := kafka.NewConsumer(kafka.ConsumerOptions{
		Brokers: settings.KafkaBrokers,
		Group:   settings.ConsumerGroup,
		Topics: []string{
			events.TopicClaimAIReviewRequested,
			events.TopicClaimAIReviewCompleted,
		},
		Logger:     logger,
		Handle:     reviews.Apply,
		DeadLetter: producer,
	})
	if err != nil {
		return err
	}
	defer consumer.Close()

	consumerCtx, stopConsumer := context.WithCancel(ctx)
	defer stopConsumer()

	go consumer.Run(consumerCtx)

	verifiers := service.NewVerifierService(store, repository.NewClaimRepository(), logger)

	sustainabilityServer := rpc.NewSustainabilityServer(store, claims, verifiers, logger, revision)

	verifier := servicetoken.NewVerifier(publicKey)
	exempt := map[string]bool{pingMethod: true}

	logger.Info("sustainability-service ready",
		slog.String("schema", settings.DatabaseSchema),
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
		ServiceName:     "carboncircuit.sustainability.v1.SustainabilityService",
		HealthInterval:  settings.HealthInterval,
		ReportHealth: func(healthCtx context.Context) bool {
			pool, poolErr := store.DB()
			return poolErr == nil && pool.PingContext(healthCtx) == nil
		},
		Register: func(server *grpc.Server) {
			sustainabilityv1.RegisterSustainabilityServiceServer(server, sustainabilityServer)
		},
	})
}
