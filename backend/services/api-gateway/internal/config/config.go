package config

import (
	"time"

	sharedconfig "github.com/carboncircuit/backend/internal/config"
	"github.com/carboncircuit/backend/internal/grpcx"
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
	ProvenanceAddress   string
	RedisAddress        string
	RedisPassword       string
	RedisDatabase       int
	PublicReadPerMinute int
	PublicReadBurst     int
	PortalUserPerMinute int
	PortalUserBurst     int
	UpstreamDialTimeout time.Duration
	UpstreamCallTimeout time.Duration

	Auth0Domain      string
	Auth0Audience    string
	ServiceTokenSeed string
	TokenLifetime    time.Duration
	CallerContextTTL time.Duration
	TLS              grpcx.TLSFiles
	KeyCacheTTL      time.Duration
	RevocationWindow time.Duration
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
		ProvenanceAddress:   loader.StringDefault("PROVENANCE_SERVICE_ADDRESS", "provenance-service:9093"),
		RedisAddress:        loader.StringDefault("REDIS_ADDRESS", "redis:6379"),
		RedisPassword:       loader.StringDefault("REDIS_PASSWORD", ""),
		RedisDatabase:       loader.Int("REDIS_DATABASE", 0),
		PublicReadPerMinute: loader.Int("PUBLIC_READ_PER_MINUTE", 60),
		PublicReadBurst:     loader.Int("PUBLIC_READ_BURST", 20),
		PortalUserPerMinute: loader.Int("PORTAL_USER_PER_MINUTE", 300),
		PortalUserBurst:     loader.Int("PORTAL_USER_BURST", 60),
		UpstreamDialTimeout: loader.Duration("UPSTREAM_DIAL_TIMEOUT", 5*time.Second),
		UpstreamCallTimeout: loader.Duration("UPSTREAM_CALL_TIMEOUT", 2*time.Second),

		Auth0Domain:      loader.String("AUTH0_DOMAIN"),
		Auth0Audience:    loader.String("AUTH0_AUDIENCE"),
		ServiceTokenSeed: loader.String("SERVICE_TOKEN_SEED"),
		TokenLifetime:    loader.Duration("SERVICE_TOKEN_LIFETIME", 30*time.Second),
		CallerContextTTL: loader.Duration("CALLER_CONTEXT_TTL", time.Minute),
		TLS: grpcx.TLSFiles{
			CertificateAuthority: loader.StringDefault("TLS_CA_FILE", ""),
			Certificate:          loader.StringDefault("TLS_CERT_FILE", ""),
			PrivateKey:           loader.StringDefault("TLS_KEY_FILE", ""),
		},
		KeyCacheTTL:      loader.Duration("AUTH0_KEY_CACHE_TTL", 5*time.Minute),
		RevocationWindow: loader.Duration("REVOCATION_WINDOW", 15*time.Minute),
	}

	return config, loader.Err()
}
