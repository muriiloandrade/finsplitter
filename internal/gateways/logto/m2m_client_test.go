package logto

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/muriiloandrade/finsplitter/pkg/httpclient"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateUser_EmailAlreadyInUse(t *testing.T) {
	// Server handles both the token endpoint (for getToken) and the
	// create-user endpoint (for CreateUser).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/token" && r.Method == http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"access_token": "test-m2m-token",
				"expires_in": 3600,
				"token_type": "Bearer"
			}`))
		case strings.HasPrefix(r.URL.Path, "/api/users") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusUnprocessableEntity)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(Config{
		OIDCEndpoint:          server.URL,
		ManagementBaseURL:     server.URL,
		ManagementAPIResource: server.URL + "/api",
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
	}, WithHTTPClientOptions(httpclient.WithRetryCount(0)))
	t.Cleanup(func() { client.httpClient.Close() })

	resp, err := client.CreateUser(context.Background(), "testuser", "", "Test", "test@example.com")

	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrEmailAlreadyInUse)
}
