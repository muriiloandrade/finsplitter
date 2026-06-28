// Package cache provides a Redis/Valkey-backed cache client with
// OpenTelemetry instrumentation, built on go-redis v9.
//
// It exposes typed methods (SetJSON, GetJSON) for struct values and GetBytes
// for raw byte access. OpenTelemetry tracing and metrics are enabled by
// default on every client.
//
// Basic usage:
//
//	c, err := cache.New(ctx, "redis://localhost:6379/0")
//	if err != nil { ... }
//	defer c.Close()
//
//	err = c.SetJSON(ctx, "mykey", myStruct, 15*time.Minute)
//	found, err := c.GetJSON(ctx, "mykey", &myStruct)
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// Client is a thin wrapper around a Redis client that supports JSON serialisation.
//
// Zero value is not usable — use New() to construct.
type Client struct {
	raw redis.UniversalClient
}

// Option is a functional option that configures a Client.
type Option func(*Client)

// New builds a Client from a Redis URL and applies the given options.
//
// The URL must be in redis:// or rediss:// format as accepted by redis.ParseURL
// (e.g. "redis://localhost:6379/0"). OpenTelemetry tracing and metrics are
// automatically enabled on every client.
//
// Callers MUST call Close() when the client is no longer needed.
func New(_ context.Context, url string, opts ...Option) (*Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("cache parse url: %w", err)
	}

	c := &Client{
		raw: redis.NewClient(opt),
	}

	for _, o := range opts {
		o(c)
	}

	if traceErr := redisotel.InstrumentTracing(c.raw); traceErr != nil {
		_ = c.raw.Close()
		return nil, fmt.Errorf("cache otel tracing: %w", traceErr)
	}
	if metricErr := redisotel.InstrumentMetrics(c.raw); metricErr != nil {
		_ = c.raw.Close()
		return nil, fmt.Errorf("cache otel metrics: %w", metricErr)
	}

	return c, nil
}

// Close releases underlying Redis connection pool resources.
func (c *Client) Close() error {
	return c.raw.Close()
}

// Ping checks connectivity to the Redis server.
func (c *Client) Ping(ctx context.Context) error {
	return c.raw.Ping(ctx).Err()
}

// SetJSON marshals val to JSON and stores it at key with the given TTL.
// A zero or negative TTL means no expiration.
func (c *Client) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}
	return c.raw.Set(ctx, key, data, ttl).Err()
}

// GetJSON retrieves the value at key and unmarshals it into dest.
// It returns (true, nil) on a cache hit, (false, nil) when the key does not
// exist, and (false, err) on any other error.
func (c *Client) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	data, err := c.raw.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache get: %w", err)
	}
	if unmarshalErr := json.Unmarshal(data, dest); unmarshalErr != nil {
		return false, fmt.Errorf("cache unmarshal: %w", unmarshalErr)
	}
	return true, nil
}

// GetBytes retrieves the raw byte slice stored at key.
// It returns (nil, nil) when the key does not exist.
func (c *Client) GetBytes(ctx context.Context, key string) ([]byte, error) {
	data, err := c.raw.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}
	return data, nil
}
