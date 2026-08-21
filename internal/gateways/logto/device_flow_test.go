package logto

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

// newTestDeviceFlowClient constructs a *Client suitable for device flow tests.
// It avoids NewClient so there is no dependency on ManagementAPIResource.
func newTestDeviceFlowClient(t *testing.T, serverURL string) *Client {
	t.Helper()

	c := &Client{
		httpClient: httpclient.New(
			httpclient.WithRetryCount(0),
			httpclient.WithTimeout(5*time.Second),
		),
		cfg: Config{
			OIDCEndpoint:      serverURL,
			DeviceAppClientID: "test-device-client",
		},
	}

	t.Cleanup(func() { c.httpClient.Close() })

	return c
}

// ---------------------------------------------------------------------------
// RequestDeviceCode
// ---------------------------------------------------------------------------

func TestClient_RequestDeviceCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/device/auth", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		_ = r.ParseForm()
		assert.Equal(t, "test-device-client", r.Form.Get("client_id"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceCodeResponse{
			DeviceCode:              "dc_123",
			UserCode:                "ABCD-EFGH",
			VerificationURI:         "http://localhost:3001/device",
			VerificationURIComplete: "http://localhost:3001/device?user_code=ABCD-EFGH",
			ExpiresIn:               1800,
			Interval:                5,
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.RequestDeviceCode(context.Background())

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "dc_123", resp.DeviceCode)
	assert.Equal(t, "ABCD-EFGH", resp.UserCode)
	assert.Equal(t, "http://localhost:3001/device", resp.VerificationURI)
	assert.Equal(t, "http://localhost:3001/device?user_code=ABCD-EFGH", resp.VerificationURIComplete)
	assert.Equal(t, 1800, resp.ExpiresIn)
	assert.Equal(t, 5, resp.Interval)
}

func TestClient_RequestDeviceCode_NotConfigured(t *testing.T) {
	client := &Client{
		httpClient: httpclient.New(httpclient.WithRetryCount(0)),
		cfg: Config{
			OIDCEndpoint: "http://fake.example.com",
			// DeviceAppClientID intentionally empty.
		},
	}
	t.Cleanup(func() { client.httpClient.Close() })

	resp, err := client.RequestDeviceCode(context.Background())

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrAppClientNotConfigured)
}

func TestClient_RequestDeviceCode_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/device/auth", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.RequestDeviceCode(context.Background())

	require.Nil(t, resp)
	require.ErrorContains(t, err, "status 500")
}

// ---------------------------------------------------------------------------
// PollDeviceToken
// ---------------------------------------------------------------------------

func TestClient_PollDeviceToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenResponse{
			AccessToken:  "access_abc",
			IDToken:      "id_def",
			RefreshToken: "refresh_ghi",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "access_abc", resp.AccessToken)
	assert.Equal(t, "id_def", resp.IDToken)
	assert.Equal(t, "refresh_ghi", resp.RefreshToken)
	assert.Equal(t, 3600, resp.ExpiresIn)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestClient_PollDeviceToken_AuthorizationPending(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenResponse{
			Error: "authorization_pending",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrDeviceCodePending)
}

func TestClient_PollDeviceToken_SlowDown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenResponse{
			Error: "slow_down",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrDeviceCodePending)
}

func TestClient_PollDeviceToken_ExpiredToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenResponse{
			Error: "expired_token",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrDeviceCodeExpired)
}

func TestClient_PollDeviceToken_AccessDenied(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenResponse{
			Error: "access_denied",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrDeviceCodeAccessDenied)
}

func TestClient_PollDeviceToken_UnknownError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenResponse{
			Error: "unknown_error",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.ErrorContains(t, err, "unknown_error")
}

func TestClient_PollDeviceToken_Non400Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.ErrorContains(t, err, "status 500")
}

func TestClient_PollDeviceToken_NotConfigured(t *testing.T) {
	client := &Client{
		httpClient: httpclient.New(httpclient.WithRetryCount(0)),
		cfg: Config{
			OIDCEndpoint: "http://fake.example.com",
		},
	}
	t.Cleanup(func() { client.httpClient.Close() })

	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrAppClientNotConfigured)
}

// ---------------------------------------------------------------------------
// RefreshDeviceToken
// ---------------------------------------------------------------------------

func TestClient_RefreshDeviceToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenRefreshResponse{
			AccessToken:  "new_access_abc",
			IDToken:      "new_id_def",
			RefreshToken: "new_refresh_ghi",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
			Scope:        "openid profile offline_access",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.RefreshDeviceToken(context.Background(), "old_refresh_token")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "new_access_abc", resp.AccessToken)
	assert.Equal(t, "new_id_def", resp.IDToken)
	assert.Equal(t, "new_refresh_ghi", resp.RefreshToken)
	assert.Equal(t, 3600, resp.ExpiresIn)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, "openid profile offline_access", resp.Scope)
}

func TestClient_RefreshDeviceToken_InvalidGrant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenRefreshResponse{
			Error: "invalid_grant",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.RefreshDeviceToken(context.Background(), "expired_token")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrDeviceCodeExpired)
}

func TestClient_RefreshDeviceToken_UnknownError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(DeviceTokenRefreshResponse{
			Error: "some_error",
		}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	resp, err := client.RefreshDeviceToken(context.Background(), "some_refresh_token")

	require.Nil(t, resp)
	require.ErrorContains(t, err, "some_error")
}

func TestClient_RefreshDeviceToken_NotConfigured(t *testing.T) {
	client := &Client{
		httpClient: httpclient.New(httpclient.WithRetryCount(0)),
		cfg: Config{
			OIDCEndpoint: "http://fake.example.com",
		},
	}
	t.Cleanup(func() { client.httpClient.Close() })

	resp, err := client.RefreshDeviceToken(context.Background(), "some_token")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrAppClientNotConfigured)
}

// ---------------------------------------------------------------------------
// RevokeDeviceToken
// ---------------------------------------------------------------------------

func TestClient_RevokeDeviceToken_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token/revocation", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		_ = r.ParseForm()
		assert.Equal(t, "test-device-client", r.Form.Get("client_id"))
		assert.Equal(t, "rt_123", r.Form.Get("token"))
		assert.Equal(t, "refresh_token", r.Form.Get("token_type_hint"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	err := client.RevokeDeviceToken(context.Background(), "rt_123")

	require.NoError(t, err)
}

func TestClient_RevokeDeviceToken_NotConfigured(t *testing.T) {
	client := &Client{
		httpClient: httpclient.New(httpclient.WithRetryCount(0)),
		cfg: Config{
			OIDCEndpoint: "http://fake.example.com",
			// DeviceAppClientID intentionally empty.
		},
	}
	t.Cleanup(func() { client.httpClient.Close() })

	err := client.RevokeDeviceToken(context.Background(), "rt_123")

	require.ErrorIs(t, err, ErrAppClientNotConfigured)
}

func TestClient_RevokeDeviceToken_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		assert.NoError(t, json.NewEncoder(w).Encode(struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}{Error: "invalid_token"}))
	}))
	defer server.Close()

	client := newTestDeviceFlowClient(t, server.URL)
	err := client.RevokeDeviceToken(context.Background(), "rt_expired")

	require.ErrorIs(t, err, ErrDeviceTokenRevoked)
	assert.Contains(t, err.Error(), "invalid_token")
}

// ---------------------------------------------------------------------------
// HTTP client error paths
// ---------------------------------------------------------------------------

func TestClient_RequestDeviceCode_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	client := newTestDeviceFlowClient(t, server.URL)
	server.Close() // Close so the next request fails with a connection error.

	resp, err := client.RequestDeviceCode(context.Background())

	require.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request device code")
}

func TestClient_PollDeviceToken_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	client := newTestDeviceFlowClient(t, server.URL)
	server.Close()

	resp, err := client.PollDeviceToken(context.Background(), "dc_123")

	require.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "poll device token")
}

func TestClient_RefreshDeviceToken_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	client := newTestDeviceFlowClient(t, server.URL)
	server.Close()

	resp, err := client.RefreshDeviceToken(context.Background(), "old_token")

	require.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "refresh device token")
}

func TestClient_RevokeDeviceToken_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	client := newTestDeviceFlowClient(t, server.URL)
	server.Close()

	err := client.RevokeDeviceToken(context.Background(), "rt_123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "revoke device token")
}
