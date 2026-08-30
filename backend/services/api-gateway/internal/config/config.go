package config

import (
	"time"

	sharedconfig "github.com/carboncircuit/backend/internal/config"
)

const ServiceName = "api-gateway"

type Config struct {
	Environment     string
	LogLevel        string
	HTTPAddress     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	IdentityAddress     string
	UpstreamDialTimeout time.Duration
	UpstreamCallTimeout time.Duration
}

func Load() (Config, error) {
	loader := sharedconfig.NewLoader(ServiceName)

	config := Config{
		Environment:     loader.StringDefault("ENVIRONMENT", "development"),
		LogLevel:        loader.StringDefault("LOG_LEVEL", "info"),
		HTTPAddress:     loader.StringDefault("HTTP_ADDRESS", ":8080"),
		ReadTimeout:     loader.Duration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    loader.Duration("WRITE_TIMEOUT", 15*time.Second),
		ShutdownTimeout: loader.Duration("SHUTDOWN_TIMEOUT", 20*time.Second),

		IdentityAddress:     loader.String("IDENTITY_SERVICE_ADDRESS"),
		UpstreamDialTimeout: loader.Duration("UPSTREAM_DIAL_TIMEOUT", 5*time.Second),
		UpstreamCallTimeout: loader.Duration("UPSTREAM_CALL_TIMEOUT", 2*time.Second),
	}

	return config, loader.Err()
}
