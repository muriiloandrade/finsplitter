package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	openapi "github.com/muriiloandrade/finsplitter/api"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TestRegisterRoutes_ContainsExpectedPaths
// ---------------------------------------------------------------------------

func TestRegisterRoutes_ContainsExpectedPaths(t *testing.T) {
	r := chi.NewRouter()
	cfg := openapi.NewOpenAPIConfig()
	api := humachi.New(r, cfg)

	a := API{
		RegisterHandler: func(_ context.Context, _ *RegisterRequest) (*RegisterResponse, error) {
			return &RegisterResponse{}, nil
		},
		MeHandler: func(_ context.Context, _ *struct{}) (*MeResponse, error) { return &MeResponse{}, nil },
		DeviceAuthHandler: func(_ context.Context, _ *RequestDeviceAuthRequest) (*RequestDeviceAuthResponse, error) {
			return &RequestDeviceAuthResponse{}, nil
		},
		DevicePollHandler: func(_ context.Context, _ *PollDeviceTokenRequest) (*PollDeviceTokenResponse, error) {
			return &PollDeviceTokenResponse{}, nil
		},
		DeviceRefreshHandler: func(_ context.Context, _ *DeviceRefreshRequest) (*DeviceRefreshResponse, error) {
			return &DeviceRefreshResponse{}, nil
		},
	}

	a.RegisterRoutes(api)

	oapi := api.OpenAPI()
	require.NotNil(t, oapi.Paths, "OpenAPI Paths must not be nil after registration")

	expectedPaths := []string{
		"/auth/register",
		"/auth/me",
		"/auth/device",
		"/auth/device/poll",
		"/auth/device/refresh",
	}

	for _, path := range expectedPaths {
		_, ok := oapi.Paths[path]
		assert.True(t, ok, "expected path %q to be registered in OpenAPI", path)
	}
}

// ---------------------------------------------------------------------------
// TestRegisterRoutes_CorrectMethods
// ---------------------------------------------------------------------------

func TestRegisterRoutes_CorrectMethods(t *testing.T) {
	r := chi.NewRouter()
	cfg := openapi.NewOpenAPIConfig()
	api := humachi.New(r, cfg)

	a := API{
		RegisterHandler: func(_ context.Context, _ *RegisterRequest) (*RegisterResponse, error) {
			return &RegisterResponse{}, nil
		},
		MeHandler: func(_ context.Context, _ *struct{}) (*MeResponse, error) { return &MeResponse{}, nil },
		DeviceAuthHandler: func(_ context.Context, _ *RequestDeviceAuthRequest) (*RequestDeviceAuthResponse, error) {
			return &RequestDeviceAuthResponse{}, nil
		},
		DevicePollHandler: func(_ context.Context, _ *PollDeviceTokenRequest) (*PollDeviceTokenResponse, error) {
			return &PollDeviceTokenResponse{}, nil
		},
		DeviceRefreshHandler: func(_ context.Context, _ *DeviceRefreshRequest) (*DeviceRefreshResponse, error) {
			return &DeviceRefreshResponse{}, nil
		},
	}

	a.RegisterRoutes(api)

	oapi := api.OpenAPI()
	require.NotNil(t, oapi.Paths, "OpenAPI Paths must not be nil after registration")

	tests := []struct {
		path           string
		expectedMethod string
	}{
		{path: "/auth/register", expectedMethod: http.MethodPost},
		{path: "/auth/me", expectedMethod: http.MethodGet},
		{path: "/auth/device", expectedMethod: http.MethodPost},
		{path: "/auth/device/poll", expectedMethod: http.MethodPost},
		{path: "/auth/device/refresh", expectedMethod: http.MethodPost},
	}

	for _, tt := range tests {
		pathItem, ok := oapi.Paths[tt.path]
		require.True(t, ok, "expected path %q to be registered", tt.path)

		var op *huma.Operation
		switch tt.expectedMethod {
		case http.MethodGet:
			op = pathItem.Get
		case http.MethodPost:
			op = pathItem.Post
		case http.MethodPut:
			op = pathItem.Put
		case http.MethodPatch:
			op = pathItem.Patch
		case http.MethodDelete:
			op = pathItem.Delete
		case http.MethodHead:
			op = pathItem.Head
		case http.MethodOptions:
			op = pathItem.Options
		case http.MethodTrace:
			op = pathItem.Trace
		default:
			t.Fatalf("unsupported HTTP method %q for path %q", tt.expectedMethod, tt.path)
		}

		require.NotNil(t, op,
			"expected operation %s %q to be registered (PathItem field is nil)",
			tt.expectedMethod, tt.path,
		)
		assert.Equal(t, tt.path, op.Path, "operation Path should match")
		assert.Equal(t, tt.expectedMethod, op.Method, "operation Method should match")
	}
}

// ---------------------------------------------------------------------------
// TestNewAPI
// ---------------------------------------------------------------------------

func TestNewAPI(t *testing.T) {
	userRepo := ports.NewMockUserRepository(t)
	logtoClient := logto.NewClient(logto.Config{
		DeviceAppClientID: "test-device",
		OIDCEndpoint:      "http://fake.example.com",
		ManagementBaseURL: "http://fake.example.com",
		ClientID:          "test-client",
		ClientSecret:      "test-secret",
	})

	api := NewAPI(userRepo, logtoClient)

	require.NotNil(t, api, "NewAPI should return a non-zero API value")
	assert.NotNil(t, api.RegisterHandler, "RegisterHandler should be wired")
	assert.NotNil(t, api.MeHandler, "MeHandler should be wired")
	assert.NotNil(t, api.DeviceAuthHandler, "DeviceAuthHandler should be wired")
	assert.NotNil(t, api.DevicePollHandler, "DevicePollHandler should be wired")
	assert.NotNil(t, api.DeviceRefreshHandler, "DeviceRefreshHandler should be wired")
}
