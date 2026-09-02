package config

import (
	"strings"
	"time"

	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/grpcx"
)

const ServiceName = "provenance-service"

type Config struct {
	ServiceTokenPublicKey string
	TLS                   grpcx.TLSFiles

	Environment     string
	LogLevel        string
	GRPCAddress     string
	ShutdownTimeout time.Duration
	HealthInterval  time.Duration

	DatabaseDSN     string
	DatabaseSchema  string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	AcquireTimeout  time.Duration

	RedisAddress  string
	RedisPassword string
	RedisDatabase int

	IdentityAddress string
	CallTimeout     time.Duration

	KafkaBrokers     []string
	KafkaTopicCreate bool
	OutboxInterval   time.Duration
	OutboxBatchSize  int
}

func Load() (Config, error) {
	loader := sharedconfig.NewLoader(ServiceName)

	config := Config{
		ServiceTokenPublicKey: loader.String("SERVICE_TOKEN_PUBLIC_KEY"),
		TLS: grpcx.TLSFiles{
			CertificateAuthority: loader.StringDefault("TLS_CA_FILE", ""),
			Certificate:          loader.StringDefault("TLS_CERT_FILE", ""),
			PrivateKey:           loader.StringDefault("TLS_KEY_FILE", ""),
		},
		Environment:     loader.StringDefault("ENVIRONMENT", "development"),
		LogLevel:        loader.StringDefault("LOG_LEVEL", "info"),
		GRPCAddress:     loader.StringDefault("GRPC_ADDRESS", ":9093"),
		ShutdownTimeout: loader.Duration("SHUTDOWN_TIMEOUT", 20*time.Second),
		HealthInterval:  loader.Duration("HEALTH_INTERVAL", 10*time.Second),

		DatabaseDSN:     loader.String("DATABASE_DSN"),
		DatabaseSchema:  loader.StringDefault("DATABASE_SCHEMA", "provenance"),
		MaxOpenConns:    loader.Int("DATABASE_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    loader.Int("DATABASE_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: loader.Duration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		ConnMaxIdleTime: loader.Duration("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute),
		AcquireTimeout:  loader.Duration("DATABASE_ACQUIRE_TIMEOUT", 250*time.Millisecond),

		RedisAddress:  loader.StringDefault("REDIS_ADDRESS", "redis:6379"),
		RedisPassword: loader.StringDefault("REDIS_PASSWORD", ""),
		RedisDatabase: loader.Int("REDIS_DATABASE", 0),

		IdentityAddress: loader.StringDefault("IDENTITY_ADDRESS", "identity-service:9091"),
		CallTimeout:     loader.Duration("CALL_TIMEOUT", 2*time.Second),

		KafkaBrokers:     strings.Split(loader.StringDefault("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaTopicCreate: loader.StringDefault("KAFKA_ALLOW_TOPIC_CREATE", "false") == "true",
		OutboxInterval:   loader.Duration("OUTBOX_INTERVAL", time.Second),
		OutboxBatchSize:  loader.Int("OUTBOX_BATCH_SIZE", 100),
	}

	return config, loader.Err()
}
