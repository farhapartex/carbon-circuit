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
	BillingAddress      string
	RedisAddress        string
	RedisPassword       string
	RedisDatabase       int
	PublicReadPerMinute int
	PublicReadBurst     int
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
		BillingAddress:      loader.String("BILLING_SERVICE_ADDRESS"),
		RedisAddress:        loader.StringDefault("REDIS_ADDRESS", "redis:6379"),
		RedisPassword:       loader.StringDefault("REDIS_PASSWORD", ""),
		RedisDatabase:       loader.Int("REDIS_DATABASE", 0),
		PublicReadPerMinute: loader.Int("PUBLIC_READ_PER_MINUTE", 60),
		PublicReadBurst:     loader.Int("PUBLIC_READ_BURST", 20),
		UpstreamDialTimeout: loader.Duration("UPSTREAM_DIAL_TIMEOUT", 5*time.Second),
		UpstreamCallTimeout: loader.Duration("UPSTREAM_CALL_TIMEOUT", 2*time.Second),
	}

	return config, loader.Err()
}
