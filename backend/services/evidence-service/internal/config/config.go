package config

import (
	"strings"
	"time"

	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/grpcx"
	"github.com/carboncircuit/backend/services/evidence-service/internal/domain"
	"github.com/carboncircuit/backend/services/evidence-service/internal/scan"
)

const ServiceName = "evidence-service"

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

	StorageEndpoint  string
	StorageAccessKey string
	StorageSecretKey string
	StorageBucket    string
	StorageUseTLS    bool
	DownloadLinkTTL  time.Duration

	ScannerKind    string
	ScannerAddress string

	MaximumByteSize int64
	MaximumPages    int

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
		GRPCAddress:     loader.StringDefault("GRPC_ADDRESS", ":9095"),
		ShutdownTimeout: loader.Duration("SHUTDOWN_TIMEOUT", 20*time.Second),
		HealthInterval:  loader.Duration("HEALTH_INTERVAL", 10*time.Second),

		DatabaseDSN:     loader.String("DATABASE_DSN"),
		DatabaseSchema:  loader.StringDefault("DATABASE_SCHEMA", "evidence"),
		MaxOpenConns:    loader.Int("DATABASE_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    loader.Int("DATABASE_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: loader.Duration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		ConnMaxIdleTime: loader.Duration("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute),
		AcquireTimeout:  loader.Duration("DATABASE_ACQUIRE_TIMEOUT", 250*time.Millisecond),

		StorageEndpoint:  loader.StringDefault("STORAGE_ENDPOINT", "minio:9000"),
		StorageAccessKey: loader.String("STORAGE_ACCESS_KEY"),
		StorageSecretKey: loader.String("STORAGE_SECRET_KEY"),
		StorageBucket:    loader.StringDefault("STORAGE_BUCKET", "carboncircuit-evidence"),
		StorageUseTLS:    loader.StringDefault("STORAGE_USE_TLS", "false") == "true",
		DownloadLinkTTL:  loader.Duration("DOWNLOAD_LINK_TTL", 5*time.Minute),

		ScannerKind:    loader.StringDefault("SCANNER_KIND", scan.Disabled),
		ScannerAddress: loader.StringDefault("SCANNER_ADDRESS", ""),

		MaximumByteSize: int64(loader.Int("MAXIMUM_BYTE_SIZE", domain.MaximumByteSize)),
		MaximumPages:    loader.Int("MAXIMUM_PAGE_COUNT", domain.MaximumPageCount),

		KafkaBrokers:     strings.Split(loader.StringDefault("KAFKA_BROKERS", "kafka:9092"), ","),
		KafkaTopicCreate: loader.StringDefault("KAFKA_ALLOW_TOPIC_CREATE", "false") == "true",
		OutboxInterval:   loader.Duration("OUTBOX_INTERVAL", time.Second),
		OutboxBatchSize:  loader.Int("OUTBOX_BATCH_SIZE", 100),
	}

	return config, loader.Err()
}

func (c Config) ScanningRequired() bool {
	return c.Environment != "development" && c.ScannerKind == scan.Disabled
}
