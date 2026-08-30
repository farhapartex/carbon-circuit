package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	redis  *redis.Client
	logger *slog.Logger
}

const operationTimeout = 100 * time.Millisecond

func New(address, password string, database int, logger *slog.Logger) *Client {
	return &Client{
		redis: redis.NewClient(&redis.Options{
			Addr:            address,
			Password:        password,
			DB:              database,
			DialTimeout:     operationTimeout,
			ReadTimeout:     operationTimeout,
			WriteTimeout:    operationTimeout,
			MaxRetries:      0,
			PoolTimeout:     operationTimeout,
			MinRetryBackoff: operationTimeout,
		}),
		logger: logger,
	}
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, operationTimeout)
}

func (c *Client) Redis() *redis.Client { return c.redis }

func (c *Client) Ping(ctx context.Context) error {
	return c.redis.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.redis.Close()
}

func (c *Client) Invalidate(ctx context.Context, key string) {
	deleteCtx, cancel := withTimeout(ctx)
	defer cancel()
	if err := c.redis.Del(deleteCtx, key).Err(); err != nil {
		c.logger.Warn("cache invalidate failed", slog.String("key", key), slog.Any("error", err))
	}
}

func ReadThrough[T any](
	ctx context.Context,
	client *Client,
	key string,
	ttl time.Duration,
	load func(context.Context) (T, error),
) (T, error) {
	var zero T

	if client != nil {
		readCtx, cancelRead := withTimeout(ctx)
		cached, err := client.redis.Get(readCtx, key).Bytes()
		cancelRead()
		switch {
		case err == nil:
			var value T
			if unmarshalErr := json.Unmarshal(cached, &value); unmarshalErr == nil {
				return value, nil
			}
			client.logger.Warn("cache entry unreadable, falling back to source",
				slog.String("key", key))
		case errors.Is(err, redis.Nil):
		default:
			client.logger.Warn("cache read failed, falling back to source",
				slog.String("key", key), slog.Any("error", err))
		}
	}

	value, err := load(ctx)
	if err != nil {
		return zero, err
	}

	if client != nil {
		encoded, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			client.logger.Warn("cache encode failed", slog.String("key", key), slog.Any("error", marshalErr))
			return value, nil
		}
		writeCtx, cancelWrite := withTimeout(ctx)
		setErr := client.redis.Set(writeCtx, key, encoded, ttl).Err()
		cancelWrite()
		if setErr != nil {
			client.logger.Warn("cache write failed", slog.String("key", key), slog.Any("error", setErr))
		}
	}

	return value, nil
}
