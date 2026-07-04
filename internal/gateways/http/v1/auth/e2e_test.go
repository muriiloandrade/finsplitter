//go:build e2e

package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	openapi "github.com/muriiloandrade/finsplitter/api"
	authHandler "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres"
	"github.com/muriiloandrade/finsplitter/internal/gateways/postgres/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcred "github.com/testcontainers/testcontainers-go/modules/redis"
)

// ---------------------------------------------------------------------------
// Shared test infrastructure (started once via TestMain)
// ---------------------------------------------------------------------------

var (
	pgDB      *testutils.TestDB
	valkeyURL string
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start PostgreSQL
	var err error
	pgDB, err = testutils.StartTestDB(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: start postgres: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := pgDB.Close(ctx); closeErr != nil {
			fmt.Fprintf(os.Stderr, "WARN: close postgres: %v\n", closeErr)
		}
	}()

	// Start Valkey (optional — middleware degrades gracefully without cache)
	valkeyContainer, vkErr := tcred.Run(ctx, "valkey/valkey:9.1.0-trixie")
	if vkErr == nil {
		valkeyURL, _ = valkeyContainer.ConnectionString(ctx)
		defer func() { _ = valkeyContainer.Terminate(ctx) }()
	} else {
		fmt.Fprintf(os.Stderr, "WARN: start valkey (tests will run without cache): %v\n", vkErr)
	}

	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Per-test environment
// ---------------------------------------------------------------------------

// e2eEnv holds the per-test environment: a running app + mock OIDC provider.
type e2eEnv struct {
	serverURL string // base URL of the Finsplitter app
	mockOIDC  *mockOIDCProvider
}

// newE2EEnv starts a mock OIDC provider, builds the Finsplitter app, and
// returns the ready-to-use environment. Cleanup is registered via t.Cleanup.
func newE2EEnv(t *testing.T) *e2eEnv {
	t.Helper()

	// Start mock OIDC provider.
	mockOIDC, err := newMockOIDCProvider("test-app-client-id")
	require.NoError(t, err, "start mock OIDC")
	t.Cleanup(mockOIDC.Close)

	// Build Finsplitter app.
	router := buildApp(t, mockOIDC)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &e2eEnv{
		serverURL: server.URL,
		mockOIDC:  mockOIDC,
	}
}

// buildApp creates the Finsplitter chi router with real handlers, real
// middleware, real PostgreSQL, and config pointing to the mock OIDC provider.
func buildApp(t *testing.T, mockOIDC *mockOIDCProvider) *chi.Mux {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// --- Database ---
	pgTxManager := &postgres.TxManager{ConnPool: pgDB.Pool()}
	userRepo := postgres.NewUserRepository(pgTxManager)

	// --- Logto client (points to mock OIDC) ---
	logtoClient := logto.NewClient(logto.Config{
		OIDCEndpoint:          mockOIDC.OIDCEndpoint(),
		ManagementBaseURL:     mockOIDC.ManagementBaseURL(),
		ManagementAPIResource: mockOIDC.ManagementBaseURL() + "/api",
		ClientID:              "test-mgmt-client-id",
		ClientSecret:          "test-mgmt-client-secret",
		AppClientID:           "test-app-client-id",
		AppClientSecret:       "test-app-client-secret",
		DeviceAppClientID:     "test-device-client-id",
	})

	// --- Auth middleware ---
	authMiddleware := authHandler.NewMiddleware(
		mockOIDC.OIDCEndpoint(), // OIDC endpoint (for JWKS URL)
		mockOIDC.OIDCEndpoint(), // expected issuer
		"test-app-client-id",    // expected audience
		userRepo,
		logger,
		nil,
	)

	// --- Auth API (handler + use cases) ---
	authAPI := authHandler.NewAPI(userRepo, logtoClient, logtoClient)

	// --- Router ---
	router := chi.NewRouter()
	router.Use(authMiddleware.Protected())

	api := humachi.New(router, openapi.NewOpenAPIConfig())
	authAPI.RegisterRoutes(api)

	return router
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// post sends a POST request and returns the raw response (body not consumed).
func (e *e2eEnv) post(t *testing.T, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, e.serverURL+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// postOK sends a POST request and returns the response body as a string.
// It asserts a 200 OK status.
func (e *e2eEnv) postOK(t *testing.T, path, body string) string {
	t.Helper()
	resp := e.post(t, path, body)
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "POST %s: %s", path, string(data))
	return string(data)
}

// get sends a GET request with an optional Bearer token and returns the raw
// response (body not consumed).
func (e *e2eEnv) get(t *testing.T, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, e.serverURL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// getOK sends a GET request and returns the response body as a string.
// It asserts a 200 OK status.
func (e *e2eEnv) getOK(t *testing.T, path, token string) string {
	t.Helper()
	resp := e.get(t, path, token)
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET %s: %s", path, string(data))
	return string(data)
}

// ---------------------------------------------------------------------------
// E2E Tests
// ---------------------------------------------------------------------------

// TestE2E_DeviceFlow_RegisterAuthMe covers the full device flow happy path:
// register → request device code → approve device code in mock → poll for
// token → use token for /auth/me.
func TestE2E_DeviceFlow_RegisterAuthMe(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	// --- Register (passwordless) ---
	regBody := env.postOK(t, "/auth/register", `{
		"name":"John Doe","email":"john@example.com"
	}`)

	var reg struct {
		UserID  string `json:"user_id"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(regBody), &reg))
	assert.NotEmpty(t, reg.UserID, "user_id")
	assert.NotEmpty(t, reg.Message, "registration message")

	// --- Request device code ---
	daBody := env.postOK(t, "/auth/device", `{
		"email":"john@example.com"
	}`)

	var da struct {
		DeviceCode              string `json:"device_code"`
		UserCode                string `json:"user_code"`
		VerificationURI         string `json:"verification_uri"`
		VerificationURIComplete string `json:"verification_uri_complete"`
		ExpiresIn               int    `json:"expires_in"`
		Interval                int    `json:"interval"`
	}
	require.NoError(t, json.Unmarshal([]byte(daBody), &da))
	assert.NotEmpty(t, da.DeviceCode, "device_code")
	assert.NotEmpty(t, da.UserCode, "user_code")
	assert.NotEmpty(t, da.VerificationURI, "verification_uri")
	assert.Greater(t, da.ExpiresIn, 0, "expires_in")
	assert.Greater(t, da.Interval, 0, "interval")

	// --- Approve the device code in the mock OIDC provider ---
	approved := env.mockOIDC.ApproveDeviceCode(da.DeviceCode, "john@example.com")
	require.True(t, approved, "approve device code")

	// --- Poll for token ---
	pollBody := env.postOK(t, "/auth/device/poll", `{
		"device_code":"`+da.DeviceCode+`"
	}`)

	var poll struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	require.NoError(t, json.Unmarshal([]byte(pollBody), &poll))
	assert.NotEmpty(t, poll.AccessToken, "access_token")
	assert.NotEmpty(t, poll.IDToken, "id_token")
	assert.NotEmpty(t, poll.RefreshToken, "refresh_token")
	assert.Greater(t, poll.ExpiresIn, 0, "expires_in")

	// --- Use token to access /auth/me ---
	meBody := env.getOK(t, "/auth/me", poll.AccessToken)

	var me struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		NeedsSetup bool   `json:"needs_setup"`
	}
	require.NoError(t, json.Unmarshal([]byte(meBody), &me))
	assert.NotEmpty(t, me.ID, "finsplitter user id")
	assert.Equal(t, "john@example.com", me.Email)
	assert.Equal(t, "John Doe", me.Name)
	assert.False(t, me.NeedsSetup, "should not need setup after register")
}

// TestE2E_DeviceFlow_PollPending verifies that polling an unapproved device
// code returns 401 (authorization_pending).
func TestE2E_DeviceFlow_PollPending(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	// Request device code.
	daBody := env.postOK(t, "/auth/device", `{"email":"user@example.com"}`)

	var da struct {
		DeviceCode string `json:"device_code"`
	}
	require.NoError(t, json.Unmarshal([]byte(daBody), &da))

	// Poll before approving — should get 401 (authorization_pending).
	resp := env.post(t, "/auth/device/poll", `{"device_code":"`+da.DeviceCode+`"}`)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "pending poll: %s", string(body))
}

// TestE2E_DeviceFlow_PollExpired verifies that polling an expired device code
// returns 400.
func TestE2E_DeviceFlow_PollExpired(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	// Request device code.
	daBody := env.postOK(t, "/auth/device", `{"email":"user@example.com"}`)

	var da struct {
		DeviceCode string `json:"device_code"`
	}
	require.NoError(t, json.Unmarshal([]byte(daBody), &da))

	// Expire the code in the mock.
	expired := env.mockOIDC.ExpireDeviceCode(da.DeviceCode)
	require.True(t, expired)

	// Poll after expiry — should get 400.
	resp := env.post(t, "/auth/device/poll", `{"device_code":"`+da.DeviceCode+`"}`)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "expired poll: %s", string(body))
}

// TestE2E_DeviceFlow_PollDenied verifies that a denied device code returns 403.
func TestE2E_DeviceFlow_PollDenied(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	// Request device code.
	daBody := env.postOK(t, "/auth/device", `{"email":"user@example.com"}`)

	var da struct {
		DeviceCode string `json:"device_code"`
	}
	require.NoError(t, json.Unmarshal([]byte(daBody), &da))

	// Deny the code in the mock.
	denied := env.mockOIDC.DenyDeviceCode(da.DeviceCode)
	require.True(t, denied)

	// Poll after denial — should get 403.
	resp := env.post(t, "/auth/device/poll", `{"device_code":"`+da.DeviceCode+`"}`)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "denied poll: %s", string(body))
}

// TestE2E_DeviceFlow_UnauthenticatedMe verifies that /auth/me works without a
// token (it's an optional-auth path) and returns an empty response.
func TestE2E_DeviceFlow_UnauthenticatedMe(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	resp := env.get(t, "/auth/me", "")
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "me: %s", string(body))

	var me map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &me))
	assert.Empty(t, me["id"], "should not have an ID without auth")
	needsSetup, _ := me["needs_setup"].(bool)
	assert.False(t, needsSetup, "needs_setup defaults to false")
}

// TestE2E_DeviceFlow_BadToken verifies that invalid tokens are silently
// ignored on the optional /auth/me path (same behavior as no token at all).
func TestE2E_DeviceFlow_BadToken(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	resp := env.get(t, "/auth/me", "this-is-not-a-valid-jwt")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "bogus token: %s", string(body))

	var result struct {
		NeedsSetup bool `json:"needs_setup"`
	}
	json.Unmarshal(body, &result)
	require.False(t, result.NeedsSetup, "bad token should yield empty response")
}

// TestE2E_DeviceFlow_InvalidDeviceCode verifies that an unknown device code
// returns a 500 (generic error from Logto client).
func TestE2E_DeviceFlow_InvalidDeviceCode(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	resp := env.post(t, "/auth/device/poll", `{"device_code":"invalid-device-code"}`)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode, "invalid code: %s", string(body))
}

// TestE2E_DeviceFlow_EmptyDeviceCode verifies that an empty device code
// returns 422.
func TestE2E_DeviceFlow_EmptyDeviceCode(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e tests not run in short mode")
	}

	env := newE2EEnv(t)

	resp := env.post(t, "/auth/device/poll", `{"device_code":""}`)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "empty code: %s", string(body))
}
