package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/muriiloandrade/finsplitter/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcred "github.com/testcontainers/testcontainers-go/modules/redis"
)

// valkeyImage matches the image used in compose.infra.yml.
const valkeyImage = "valkey/valkey:9.1.0-trixie"

// newTestJWKS generates an ephemeral EC P-256 key and returns it as a jwk.Set
// suitable for caching tests. The key is not registered with any CA — it is
// valid only within the test process.
func newTestJWKS(t *testing.T) jwk.Set {
	t.Helper()

	rawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "GenerateKey")

	jwkKey, err := jwk.Import[jwk.Key](rawKey)
	require.NoError(t, err, "jwk.Import")

	keyJSON, err := json.Marshal(jwkKey)
	require.NoError(t, err, "json.Marshal(jwkKey)")

	setJSON := fmt.Sprintf(`{"keys":[%s]}`, keyJSON)
	set, err := jwk.Parse([]byte(setJSON))
	require.NoError(t, err, "jwk.Parse")

	return set
}

// newTestCacheClient starts a Valkey container via testcontainers and returns
// a cache.Client connected to it. The container is terminated when the test
// completes.
func newTestCacheClient(t *testing.T) *cache.Client {
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

// newTestMiddleware builds a Middleware with a mock JWKS fetcher and a real
// Valkey-backed cache. It returns the middleware, the mock fetcher (so the
// caller can set expectations), and the cache client.
func newTestMiddleware(t *testing.T) (*Middleware, *mockjwkFetcher, *cache.Client) {
	t.Helper()

	cacheClient := newTestCacheClient(t)
	mockFetcher := newMockjwkFetcher(t)

	mw := &Middleware{
		jwkClient: mockFetcher,
		cache:     cacheClient,
		logger:    slog.Default(),
		jwksURL:   "http://test.local/jwks",
		issuer:    "http://test.local",
	}

	return mw, mockFetcher, cacheClient
}

func TestGetJWKS_CacheHit(t *testing.T) {
	mw, mockFetcher, cacheClient := newTestMiddleware(t)
	ctx := context.Background()

	// Seed the cache with a test JWKS.
	seed := newTestJWKS(t)
	err := cacheClient.SetJSON(ctx, jwksCacheKey, seed, jwksCacheTTL)
	require.NoError(t, err, "seed cache")

	// getJWKS should return the cached set without calling the fetcher.
	got, err := mw.getJWKS(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got, "getJWKS should return a non-nil set")

	// The mock should NOT have been called — cache hit.
	mockFetcher.AssertNotCalled(t, "Fetch")
}

func TestGetJWKS_CacheMissThenHit(t *testing.T) {
	mw, mockFetcher, _ := newTestMiddleware(t)
	ctx := context.Background()

	expected := newTestJWKS(t)

	// First call: cache miss → fetcher returns the JWKS.
	mockFetcher.EXPECT().Fetch(ctx, "http://test.local/jwks").Return(expected, nil)

	got1, err := mw.getJWKS(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got1, "first call should return JWKS")

	// Second call: should be a cache hit (fetcher not called again).
	got2, err := mw.getJWKS(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got2, "second call should return JWKS")

	// Fetch should have been called exactly once.
	mockFetcher.AssertNumberOfCalls(t, "Fetch", 1)
}

func TestGetJWKS_NoCache(t *testing.T) {
	// Middleware without a cache → every call goes to the fetcher.
	mockFetcher := newMockjwkFetcher(t)
	mw := &Middleware{
		jwkClient: mockFetcher,
		cache:     nil,
		logger:    slog.Default(),
		jwksURL:   "http://test.local/jwks",
		issuer:    "http://test.local",
	}
	ctx := context.Background()

	expected := newTestJWKS(t)
	mockFetcher.EXPECT().Fetch(ctx, "http://test.local/jwks").Return(expected, nil).Twice()

	got1, err := mw.getJWKS(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got1)

	got2, err := mw.getJWKS(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got2)

	mockFetcher.AssertNumberOfCalls(t, "Fetch", 2)
}

func TestGetJWKS_RedisUnreachable(t *testing.T) {
	// A cache client pointing at a port where nothing listens → GetBytes
	// returns an error → getJWKS falls through to the fetcher.
	badCache, err := cache.New(context.Background(), "redis://localhost:16379/0")
	require.NoError(t, err, "creating cache with bad port should not fail (lazy conn)")

	mockFetcher := newMockjwkFetcher(t)
	mw := &Middleware{
		jwkClient: mockFetcher,
		cache:     badCache,
		logger:    slog.Default(),
		jwksURL:   "http://test.local/jwks",
		issuer:    "http://test.local",
	}
	ctx := context.Background()

	expected := newTestJWKS(t)
	mockFetcher.EXPECT().Fetch(ctx, "http://test.local/jwks").Return(expected, nil)

	got, err := mw.getJWKS(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got)

	mockFetcher.AssertNumberOfCalls(t, "Fetch", 1)
	_ = badCache.Close()
}

func TestGetJWKS_FetcherError(t *testing.T) {
	mw, mockFetcher, _ := newTestMiddleware(t)
	ctx := context.Background()

	// Both cache and fetcher fail → getJWKS returns the fetcher error.
	expectedErr := errors.New("logto unavailable")
	mockFetcher.EXPECT().Fetch(ctx, "http://test.local/jwks").Return(nil, expectedErr)

	got, err := mw.getJWKS(ctx)
	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, got)
}
