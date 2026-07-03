package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// mockUniversalClient
// ---------------------------------------------------------------------------

type mockUniversalClient struct {
	redis.UniversalClient // embedded nil — satisfies the interface at compile time

	closeFn func() error
	pingFn  func(ctx context.Context) *redis.StatusCmd
	setFn   func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	getFn   func(ctx context.Context, key string) *redis.StringCmd
}

func (m *mockUniversalClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func (m *mockUniversalClient) Ping(ctx context.Context) *redis.StatusCmd {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return redis.NewStatusResult("PONG", nil)
}

func (m *mockUniversalClient) Set(
	ctx context.Context,
	key string,
	value any,
	expiration time.Duration,
) *redis.StatusCmd {
	if m.setFn != nil {
		return m.setFn(ctx, key, value, expiration)
	}
	return redis.NewStatusResult("OK", nil)
}

func (m *mockUniversalClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if m.getFn != nil {
		return m.getFn(ctx, key)
	}
	return redis.NewStringResult("", nil)
}

// ---------------------------------------------------------------------------
// SetJSON
// ---------------------------------------------------------------------------

func TestCache_SetJSON_RedisError(t *testing.T) {
	mock := &mockUniversalClient{
		setFn: func(_ context.Context, _ string, _ any, _ time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("", errors.New("connection refused"))
		},
	}
	c := &Client{raw: mock}
	ctx := context.Background()

	err := c.SetJSON(ctx, "key", "val", 0)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// ---------------------------------------------------------------------------
// GetJSON
// ---------------------------------------------------------------------------

func TestCache_GetJSON_RedisGetError(t *testing.T) {
	mock := &mockUniversalClient{
		getFn: func(_ context.Context, _ string) *redis.StringCmd {
			return redis.NewStringResult("", errors.New("connection refused"))
		},
	}
	c := &Client{raw: mock}
	ctx := context.Background()

	var dest string
	found, err := c.GetJSON(ctx, "key", &dest)

	require.Error(t, err)
	assert.False(t, found)
	assert.Contains(t, err.Error(), "cache get")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestCache_GetJSON_UnmarshalError(t *testing.T) {
	mock := &mockUniversalClient{
		getFn: func(_ context.Context, _ string) *redis.StringCmd {
			return redis.NewStringResult("not-json", nil)
		},
	}
	c := &Client{raw: mock}
	ctx := context.Background()

	var dest string
	found, err := c.GetJSON(ctx, "key", &dest)

	require.Error(t, err)
	assert.False(t, found)
	assert.Contains(t, err.Error(), "cache unmarshal")
}

// ---------------------------------------------------------------------------
// GetBytes
// ---------------------------------------------------------------------------

func TestCache_GetBytes_RedisError(t *testing.T) {
	mock := &mockUniversalClient{
		getFn: func(_ context.Context, _ string) *redis.StringCmd {
			return redis.NewStringResult("", errors.New("connection closed"))
		},
	}
	c := &Client{raw: mock}
	ctx := context.Background()

	data, err := c.GetBytes(ctx, "key")

	require.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "connection closed")
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

func TestCache_Ping_RedisError(t *testing.T) {
	mock := &mockUniversalClient{
		pingFn: func(_ context.Context) *redis.StatusCmd {
			return redis.NewStatusResult("", errors.New("not connected"))
		},
	}
	c := &Client{raw: mock}
	ctx := context.Background()

	err := c.Ping(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// ---------------------------------------------------------------------------
// Close
// ---------------------------------------------------------------------------

func TestCache_Close_Error(t *testing.T) {
	mock := &mockUniversalClient{
		closeFn: func() error {
			return errors.New("close error")
		},
	}
	c := &Client{raw: mock}

	err := c.Close()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "close error")
}
