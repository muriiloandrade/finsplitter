package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/muriiloandrade/finsplitter/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcred "github.com/testcontainers/testcontainers-go/modules/redis"
)

// valkeyImage matches the image used in compose.infra.yml.
const valkeyImage = "valkey/valkey:9.1.0-trixie"

// newTestClient starts a Valkey container via testcontainers and returns a
// cache.Client connected to it. The container is terminated when the test
// completes.
func newTestClient(t *testing.T) *cache.Client {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping testcontainers test in short mode")
	}

	ctx := context.Background()

	container, err := tcred.Run(ctx, valkeyImage)
	require.NoError(t, err, "Failed to start Valkey container")

	url, err := container.ConnectionString(ctx)
	require.NoError(t, err, "Failed to get connection string")

	client, err := cache.New(ctx, url)
	require.NoError(t, err, "Failed to create cache client")

	t.Cleanup(func() {
		_ = client.Close()
		_ = container.Terminate(ctx)
	})

	return client
}

func TestSetGetJSON_Roundtrip(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	key := "test:roundtrip"

	type widget struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	original := widget{ID: 42, Name: "the-answer"}
	err := client.SetJSON(ctx, key, original, 5*time.Minute)
	require.NoError(t, err, "SetJSON should succeed")

	var decoded widget
	found, err := client.GetJSON(ctx, key, &decoded)
	require.NoError(t, err, "GetJSON should succeed")
	assert.True(t, found, "key should exist")
	assert.Equal(t, original, decoded, "roundtrip value should match")
}

func TestGetJSON_KeyNotFound(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	var dest struct{}
	found, err := client.GetJSON(ctx, "test:nonexistent", &dest)
	require.NoError(t, err, "GetJSON for missing key should not error")
	assert.False(t, found, "found should be false for missing key")
}

func TestGetJSON_TTLExpiration(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	key := "test:ttl"

	err := client.SetJSON(ctx, key, "ephemeral", 1*time.Second)
	require.NoError(t, err, "SetJSON should succeed")

	// Read before expiry — must be present.
	var got string
	found, err := client.GetJSON(ctx, key, &got)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "ephemeral", got)

	// Wait for TTL to expire.
	time.Sleep(1100 * time.Millisecond)

	found, err = client.GetJSON(ctx, key, &got)
	require.NoError(t, err, "GetJSON after expiry should not error")
	assert.False(t, found, "key should have expired")
}

func TestGetBytes_RawAccess(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	key := "test:raw"

	err := client.SetJSON(ctx, key, map[string]int{"a": 1}, 5*time.Minute)
	require.NoError(t, err)

	data, err := client.GetBytes(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, data, "GetBytes should return data for existing key")
	assert.JSONEq(t, `{"a":1}`, string(data), "raw bytes should be valid JSON")
}

func TestGetBytes_KeyNotFound(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	data, err := client.GetBytes(ctx, "test:nonexistent")
	require.NoError(t, err, "GetBytes for missing key should not error")
	assert.Nil(t, data, "data should be nil for missing key")
}

func TestConcurrentAccess(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	done := make(chan struct{})
	const workers = 10

	for range workers {
		go func() {
			defer func() { done <- struct{}{} }()
			for i := range 100 {
				key := "test:concurrent"
				_ = client.SetJSON(ctx, key, i, 30*time.Second)
				var got int
				_, _ = client.GetJSON(ctx, key, &got)
			}
		}()
	}

	for range workers {
		<-done
	}

	// Final read to confirm no corruption.
	var final int
	found, err := client.GetJSON(ctx, "test:concurrent", &final)
	require.NoError(t, err)
	assert.True(t, found)
}
