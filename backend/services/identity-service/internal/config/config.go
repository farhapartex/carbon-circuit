package config

import (
	"time"

	sharedconfig "github.com/carboncircuit/backend/internal/config"
)

const ServiceName = "identity-service"

type Config struct {
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
}

func Load() (Config, error) {
	loader := sharedconfig.NewLoader(ServiceName)

	config := Config{
		Environment:     loader.StringDefault("ENVIRONMENT", "development"),
		LogLevel:        loader.StringDefault("LOG_LEVEL", "info"),
		GRPCAddress:     loader.StringDefault("GRPC_ADDRESS", ":9091"),
		ShutdownTimeout: loader.Duration("SHUTDOWN_TIMEOUT", 20*time.Second),
		HealthInterval:  loader.Duration("HEALTH_INTERVAL", 10*time.Second),

		DatabaseDSN:     loader.String("DATABASE_DSN"),
		DatabaseSchema:  loader.StringDefault("DATABASE_SCHEMA", "identity"),
		MaxOpenConns:    loader.Int("DATABASE_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    loader.Int("DATABASE_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: loader.Duration("DATABASE_CONN_MAX_LIFETIME", 30*time.Minute),
		ConnMaxIdleTime: loader.Duration("DATABASE_CONN_MAX_IDLE_TIME", 5*time.Minute),
		AcquireTimeout:  loader.Duration("DATABASE_ACQUIRE_TIMEOUT", 250*time.Millisecond),
	}

	return config, loader.Err()
}
