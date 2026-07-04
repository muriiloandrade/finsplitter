package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/muriiloandrade/finsplitter/pkg/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ────────────────────────────────────────────────────────────────────────────
// userInfoClient
// ────────────────────────────────────────────────────────────────────────────

func TestUserInfoClient_ReturnsSetClient(t *testing.T) {
	hc := httpclient.New(
		httpclient.WithRetryCount(0),
		httpclient.WithTimeout(5*time.Second),
	)
	mw := &Middleware{
		httpClient: hc,
	}

	got := mw.userInfoClient()
	assert.Same(t, hc, got, "should return the same *httpclient.Client")
	t.Cleanup(hc.Close)
}

func TestUserInfoClient_FallbackWhenNil(t *testing.T) {
	mw := &Middleware{
		httpClient: nil,
	}

	got := mw.userInfoClient()
	require.NotNil(t, got, "should return a non-nil *httpclient.Client")
}

// ────────────────────────────────────────────────────────────────────────────
// parseViaUserInfo
// ────────────────────────────────────────────────────────────────────────────

func TestParseViaUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer test-opaque-token", r.Header.Get("Authorization"))

		resp := userInfoResponse{
			Sub:      "logto_sub_1",
			Email:    "user@example.com",
			Name:     "John",
			Username: "john",
			Phone:    "+123",
			Picture:  "https://example.com/avatar.png",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	mw := &Middleware{
		userInfoURL: server.URL,
		httpClient: httpclient.New(
			httpclient.WithRetryCount(0),
			httpclient.WithTimeout(5*time.Second),
		),
	}
	t.Cleanup(mw.httpClient.Close)

	claims, err := mw.parseViaUserInfo(context.Background(), "test-opaque-token")
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, "logto_sub_1", claims.Sub)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "John", claims.Name)
	assert.Equal(t, "john", claims.Username)
	assert.Equal(t, "+123", claims.Phone)
	assert.Equal(t, "https://example.com/avatar.png", claims.Picture)
}

func TestParseViaUserInfo_MissingSub(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		resp := userInfoResponse{
			Email: "user@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	mw := &Middleware{
		userInfoURL: server.URL,
		httpClient: httpclient.New(
			httpclient.WithRetryCount(0),
			httpclient.WithTimeout(5*time.Second),
		),
	}
	t.Cleanup(mw.httpClient.Close)

	claims, err := mw.parseViaUserInfo(context.Background(), "test-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing sub claim")
	assert.Nil(t, claims)
}

func TestParseViaUserInfo_Non200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid token"}`))
	}))
	t.Cleanup(server.Close)

	mw := &Middleware{
		userInfoURL: server.URL,
		httpClient: httpclient.New(
			httpclient.WithRetryCount(0),
			httpclient.WithTimeout(5*time.Second),
		),
	}
	t.Cleanup(mw.httpClient.Close)

	claims, err := mw.parseViaUserInfo(context.Background(), "test-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 401")
	assert.Nil(t, claims)
}

func TestParseViaUserInfo_EndpointNotConfigured(t *testing.T) {
	mw := &Middleware{
		userInfoURL: "",
		httpClient: httpclient.New(
			httpclient.WithRetryCount(0),
			httpclient.WithTimeout(5*time.Second),
		),
	}
	t.Cleanup(mw.httpClient.Close)

	claims, err := mw.parseViaUserInfo(context.Background(), "test-token")
	require.Error(t, err)
	assert.Equal(t, "userinfo endpoint not configured", err.Error())
	assert.Nil(t, claims)
}

func TestParseViaUserInfo_NetworkError(t *testing.T) {
	mw := &Middleware{
		// Port 1 is a privileged port unlikely to be listening → immediate
		// connection refused on most systems.
		userInfoURL: "http://127.0.0.1:1",
		httpClient: httpclient.New(
			httpclient.WithRetryCount(0),
			httpclient.WithTimeout(5*time.Second),
		),
	}
	t.Cleanup(mw.httpClient.Close)

	claims, err := mw.parseViaUserInfo(context.Background(), "test-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, claims)

	// Verify the error is wrapped with the userinfo prefix.
	assert.Contains(t, err.Error(), "userinfo call:")
}

// ────────────────────────────────────────────────────────────────────────────
// parseViaUserInfo — fallback client (nil httpClient)
// ────────────────────────────────────────────────────────────────────────────

func TestParseViaUserInfo_WithFallbackClient(t *testing.T) {
	// When httpClient is nil, userInfoClient() creates a one-off client.
	// This test verifies the fallback client works end-to-end.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "Bearer fallback-token", r.Header.Get("Authorization"))

		resp := userInfoResponse{
			Sub:   "fallback-user",
			Email: "fallback@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	mw := &Middleware{
		userInfoURL: server.URL,
		httpClient:  nil, // triggers fallback in userInfoClient()
	}

	claims, err := mw.parseViaUserInfo(context.Background(), "fallback-token")
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "fallback-user", claims.Sub)
	assert.Equal(t, "fallback@example.com", claims.Email)
}
