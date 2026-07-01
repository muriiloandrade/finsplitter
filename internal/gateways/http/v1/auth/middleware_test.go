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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// ────────────────────────────────────────────────────────────────────────────
// Path matching helpers
// ────────────────────────────────────────────────────────────────────────────

func TestIsPublicPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		// Prefix matches
		{"health with trailing slash", "/health/", true},
		{"health sub-path", "/health/status", true},
		{"docs prefix", "/docs", true},
		{"docs sub-path", "/docs/v1/api", true},
		{"openapi prefix", "/openapi.json", true},
		{"openapi sub-path", "/openapi/v3", true},

		// Exact match
		{"register exact", "/auth/register", true},

		// Should not match
		{"health without trailing slash", "/health", false},
		{"health with suffix", "/healthx", false},
		{"docs with suffix (prefix match)", "/docsx", true},
		{"register sub-path", "/auth/register/extra", false},
		{"protected endpoint", "/api/users", false},
		{"optional endpoint", "/auth/me", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isPublicPath(tt.path))
		})
	}
}

func TestIsOptionalPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"me exact", "/auth/me", true},
		{"setup exact", "/profile/setup", true},
		{"me with trailing slash", "/auth/me/", false},
		{"me sub-path", "/auth/me/details", false},
		{"setup sub-path", "/profile/setup/extra", false},
		{"public endpoint", "/auth/register", false},
		{"health prefix", "/health/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isOptionalPath(tt.path))
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Cache corruption — graceful degradation
// ────────────────────────────────────────────────────────────────────────────

func TestGetJWKS_CacheCorrupt(t *testing.T) {
	mw, mockFetcher, cacheClient := newTestMiddleware(t)
	ctx := context.Background()

	// Seed the cache with data that is valid JSON but not a valid JWKS.
	err := cacheClient.SetJSON(ctx, jwksCacheKey, "this is not a jwks", jwksCacheTTL)
	require.NoError(t, err)

	// getJWKS should try the cache, fail to parse, log a warning, and fall
	// through to the fetcher.
	expected := newTestJWKS(t)
	mockFetcher.EXPECT().Fetch(ctx, mw.jwksURL).Return(expected, nil).Once()

	got, err := mw.getJWKS(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got)
	mockFetcher.AssertNumberOfCalls(t, "Fetch", 1)
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers for requireAuth / tryPopulateClaims HTTP tests
// ────────────────────────────────────────────────────────────────────────────

// newTestKeySet generates an EC P-256 key pair and returns a JWK signing key
// (with kid set) together with a public JWKS (for verification). The kid is
// set on both the signing key and the JWKS so that jwt.WithKeySet can match
// them (kid matching is required by default in jwx v4).
func newTestKeySet(t *testing.T) (jwk.Key, jwk.Set) {
	t.Helper()

	rawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Signing key — kid and algorithm are included in the JWS header.
	signingKey, err := jwk.Import[jwk.Key](rawKey)
	require.NoError(t, err)
	require.NoError(t, signingKey.Set(jwk.KeyIDKey, "test-key-1"))
	require.NoError(t, signingKey.Set(jwk.AlgorithmKey, jwa.ES256()))

	// Public key in the JWKS must have the same kid and algorithm.
	pubJWK, err := jwk.Import[jwk.Key](&rawKey.PublicKey)
	require.NoError(t, err)
	require.NoError(t, pubJWK.Set(jwk.KeyIDKey, "test-key-1"))
	require.NoError(t, pubJWK.Set(jwk.AlgorithmKey, jwa.ES256()))

	keyJSON, err := json.Marshal(pubJWK)
	require.NoError(t, err)

	setJSON := fmt.Sprintf(`{"keys":[%s]}`, keyJSON)
	set, err := jwk.Parse([]byte(setJSON))
	require.NoError(t, err)

	return signingKey, set
}

// newTestSignedToken builds a JWT with the given claims and signs it with the
// provided JWK key using ES256. The key must have a kid set (use newTestKeySet).
func newTestSignedToken(t *testing.T, key jwk.Key, claims map[string]any) string {
	t.Helper()

	builder := jwt.NewBuilder()
	for k, v := range claims {
		builder = builder.Claim(k, v)
	}
	tok, err := builder.Build()
	require.NoError(t, err)

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES256(), key))
	require.NoError(t, err)

	return string(signed)
}

// captureHandler is an http.HandlerFunc that records the UserClaims from
// context (if any) into response headers and always writes 200 OK.
func captureHandler(w http.ResponseWriter, r *http.Request) {
	claims := GetUserClaims(r.Context())
	if claims != nil {
		w.Header().Set("X-Claims-Sub", claims.Sub)
		w.Header().Set("X-Claims-Email", claims.Email)
		w.Header().Set("X-Claims-Username", claims.Username)
	}
	w.WriteHeader(http.StatusOK)
}

// Test RequireAuth — no token → 401
func TestRequireAuth_NoToken(t *testing.T) {
	mw := &Middleware{
		logger: slog.Default(),
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "missing authorization header", body["error"])
}

// Test RequireAuth — invalid JWT → 401
func TestRequireAuth_InvalidToken(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	_, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient: mockFetcher,
		cache:     nil,
		logger:    slog.Default(),
		jwksURL:   "http://test.local/jwks",
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	// getJWKS must succeed so parseAndValidate can try jwt.Parse on the
	// garbage token — jwt.Parse will fail regardless of the keyset.
	// The context is wrapped by slogctx.NewCtx in requireAuth, so we use
	// mock.Anything to match the wrapped value.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.jwt")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// Test RequireAuth — valid JWT but user does not exist → 403
func TestRequireAuth_UserNotFound(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	userRepo := ports.NewMockUserRepository(t)
	privateKey, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient:   mockFetcher,
		cache:       nil,
		userRepo:    userRepo,
		logger:      slog.Default(),
		jwksURL:     "http://test.local/jwks",
		issuer:      "http://test.local",
		appClientID: "test-client",
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	// The context is wrapped by slogctx.NewCtx in requireAuth, so we use
	// mock.Anything to match the wrapped value.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()
	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "user-123").
		Return(false, nil).
		Once()

	token := newTestSignedToken(t, privateKey, map[string]any{
		"sub": "user-123",
		"iss": mw.issuer,
		"aud": mw.appClientID,
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// Test RequireAuth — valid JWT but DB error → 500
func TestRequireAuth_DBError(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	userRepo := ports.NewMockUserRepository(t)
	privateKey, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient:   mockFetcher,
		cache:       nil,
		userRepo:    userRepo,
		logger:      slog.Default(),
		jwksURL:     "http://test.local/jwks",
		issuer:      "http://test.local",
		appClientID: "test-client",
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	// The context is wrapped by slogctx.NewCtx in requireAuth, so we use
	// mock.Anything to match the wrapped value.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()
	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "user-123").
		Return(false, errors.New("db unreachable")).
		Once()

	token := newTestSignedToken(t, privateKey, map[string]any{
		"sub": "user-123",
		"iss": mw.issuer,
		"aud": mw.appClientID,
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// Test RequireAuth — valid JWT + user exists → 200 + claims in response
func TestRequireAuth_Success(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	userRepo := ports.NewMockUserRepository(t)
	privateKey, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient:   mockFetcher,
		cache:       nil,
		userRepo:    userRepo,
		logger:      slog.Default(),
		jwksURL:     "http://test.local/jwks",
		issuer:      "http://test.local",
		appClientID: "test-client",
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	// The context is wrapped by slogctx.NewCtx in requireAuth, so we use
	// mock.Anything to match the wrapped value.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()
	userRepo.EXPECT().
		ExistsByLogtoUserID(mock.Anything, "user-123").
		Return(true, nil).
		Once()

	token := newTestSignedToken(t, privateKey, map[string]any{
		"sub":      "user-123",
		"iss":      mw.issuer,
		"aud":      mw.appClientID,
		"username": "johndoe",
		"email":    "john@example.com",
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-123", rec.Header().Get("X-Claims-Sub"))
	assert.Equal(t, "john@example.com", rec.Header().Get("X-Claims-Email"))
	assert.Equal(t, "johndoe", rec.Header().Get("X-Claims-Username"))
}

// Test Protected — public path skips authentication entirely.
func TestProtected_PublicPath(t *testing.T) {
	mw := &Middleware{
		logger: slog.Default(),
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("X-Claims-Sub"),
		"public path should NOT populate claims")
}

// Test Protected — optional path populates claims when a valid token is sent.
func TestProtected_OptionalPath_WithValidToken(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	userRepo := ports.NewMockUserRepository(t)
	privateKey, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient:   mockFetcher,
		cache:       nil,
		userRepo:    userRepo,
		logger:      slog.Default(),
		jwksURL:     "http://test.local/jwks",
		issuer:      "http://test.local",
		appClientID: "test-client",
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	// The context is wrapped by slogctx.NewCtx in tryPopulateClaims, so we use
	// mock.Anything to match the wrapped value.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()

	token := newTestSignedToken(t, privateKey, map[string]any{
		"sub": "user-123",
		"iss": mw.issuer,
		"aud": mw.appClientID,
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-123", rec.Header().Get("X-Claims-Sub"),
		"optional path should populate claims with a valid token")
}

// Test Protected — optional path without token still passes.
func TestProtected_OptionalPath_NoToken(t *testing.T) {
	mw := &Middleware{
		logger: slog.Default(),
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("X-Claims-Sub"),
		"optional path without token should NOT populate claims")
}

// Test Protected — setup path with valid token passes middleware even without
// a local DB record. This is the recovery scenario: user exists in Logto but
// the previous DB insert failed (or setup was never completed).
func TestProtected_SetupPath_WithValidToken(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	privateKey, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient:   mockFetcher,
		cache:       nil,
		logger:      slog.Default(),
		jwksURL:     "http://test.local/jwks",
		issuer:      "http://test.local",
		appClientID: "test-client",
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	// No userRepo mock needed — optional-auth paths do not check DB existence.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()

	token := newTestSignedToken(t, privateKey, map[string]any{
		"sub":      "logto-user-789",
		"iss":      mw.issuer,
		"aud":      mw.appClientID,
		"username": "setup-user",
		"email":    "setup@example.com",
	})

	req := httptest.NewRequest(http.MethodPost, "/profile/setup", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "logto-user-789", rec.Header().Get("X-Claims-Sub"))
}

// Test Protected — setup path without token still passes through middleware
// (handler will return 401 since no claims are available).
func TestProtected_SetupPath_NoToken(t *testing.T) {
	mw := &Middleware{
		logger: slog.Default(),
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	req := httptest.NewRequest(http.MethodPost, "/profile/setup", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("X-Claims-Sub"),
		"setup path without token should NOT populate claims")
}

// Test Protected — protected path without token returns 401.
func TestProtected_ProtectedPath(t *testing.T) {
	mw := &Middleware{
		logger: slog.Default(),
	}
	handler := mw.Protected()(http.HandlerFunc(captureHandler))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ────────────────────────────────────────────────────────────────────────────
// tryPopulateClaims
// ────────────────────────────────────────────────────────────────────────────

func TestTryPopulateClaims_NoToken(t *testing.T) {
	mw := &Middleware{
		logger: slog.Default(),
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	modified := mw.tryPopulateClaims(req)
	assert.Nil(t, modified, "no token → should return nil")
}

func TestTryPopulateClaims_InvalidToken(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	_, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient: mockFetcher,
		cache:     nil,
		logger:    slog.Default(),
		jwksURL:   "http://test.local/jwks",
	}
	// The context is wrapped by slogctx.NewCtx in tryPopulateClaims, so we use
	// mock.Anything to match the wrapped value.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer clearly-invalid")
	modified := mw.tryPopulateClaims(req)
	assert.Nil(t, modified, "invalid token → should return nil")
}

func TestTryPopulateClaims_ValidToken(t *testing.T) {
	mockFetcher := newMockjwkFetcher(t)
	privateKey, jwks := newTestKeySet(t)

	mw := &Middleware{
		jwkClient:   mockFetcher,
		cache:       nil,
		logger:      slog.Default(),
		jwksURL:     "http://test.local/jwks",
		issuer:      "http://test.local",
		appClientID: "test-client",
	}

	// The context is wrapped by slogctx.NewCtx in tryPopulateClaims, so we use
	// mock.Anything to match the wrapped value.
	mockFetcher.EXPECT().Fetch(mock.Anything, mw.jwksURL).Return(jwks, nil).Once()

	token := newTestSignedToken(t, privateKey, map[string]any{
		"sub":      "user-456",
		"iss":      mw.issuer,
		"aud":      mw.appClientID,
		"username": "alice",
		"email":    "alice@example.com",
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	modified := mw.tryPopulateClaims(req)
	require.NotNil(t, modified, "valid token → should return modified request")

	claims := GetUserClaims(modified.Context())
	require.NotNil(t, claims)
	assert.Equal(t, "user-456", claims.Sub)
	assert.Equal(t, "alice", claims.Username)
	assert.Equal(t, "alice@example.com", claims.Email)
}
