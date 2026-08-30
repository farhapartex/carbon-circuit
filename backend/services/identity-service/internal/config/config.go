package config

import (
	"time"

	sharedconfig "github.com/carboncircuit/backend/internal/config"
)

const ServiceName = "identity-service"

type Config struct {
	Environment     string
	LogLevel        string
	HTTPAddress     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

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
		HTTPAddress:     loader.StringDefault("HTTP_ADDRESS", ":8081"),
		ReadTimeout:     loader.Duration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    loader.Duration("WRITE_TIMEOUT", 15*time.Second),
		ShutdownTimeout: loader.Duration("SHUTDOWN_TIMEOUT", 20*time.Second),

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
