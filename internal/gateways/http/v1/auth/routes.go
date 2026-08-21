package auth

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// HumaHandler defines the function signature for Huma handlers.
type HumaHandler[I, O any] func(context.Context, *I) (*O, error)

// tagAuth is the OpenAPI tag shared by all auth operations.
const tagAuth = "Auth"

// API holds the auth handler references and registers routes.
type API struct {
	RegisterHandler      HumaHandler[RegisterRequest, RegisterResponse]
	MeHandler            HumaHandler[struct{}, MeResponse]
	DeviceAuthHandler    HumaHandler[RequestDeviceAuthRequest, RequestDeviceAuthResponse]
	DevicePollHandler    HumaHandler[PollDeviceTokenRequest, PollDeviceTokenResponse]
	DeviceRefreshHandler HumaHandler[DeviceRefreshRequest, DeviceRefreshResponse]
	DeviceRevokeHandler  HumaHandler[DeviceRevokeRequest, DeviceRevokeResponse]
}

// NewAPI creates an auth API from the given dependencies.
// logtoClient satisfies both LogtoUserCreator (for registration) and
// LogtoDeviceFlowClient (for device auth) via compile-time interface checks.
func NewAPI(userRepo ports.UserRepository, logtoClient *logto.Client) API {
	registerUC := auth.NewRegisterUseCase(userRepo, logtoClient)
	meUC := auth.NewMeUseCase(userRepo)
	deviceAuthUC := auth.NewRequestDeviceAuthUseCase(logtoClient)
	devicePollUC := auth.NewPollDeviceTokenUseCase(logtoClient)
	deviceRefreshUC := auth.NewRefreshDeviceTokenUseCase(logtoClient)
	deviceRevokeUC := auth.NewRevokeDeviceTokenUseCase(logtoClient)
	h := NewHandler(HandlerConfig{
		RegisterUC:      registerUC,
		MeUC:            meUC,
		DeviceAuthUC:    deviceAuthUC,
		DevicePollUC:    devicePollUC,
		DeviceRefreshUC: deviceRefreshUC,
		DeviceRevokeUC:  deviceRevokeUC,
	})

	return API{
		RegisterHandler:      h.Register,
		MeHandler:            h.Me,
		DeviceAuthHandler:    h.DeviceAuth,
		DevicePollHandler:    h.DevicePoll,
		DeviceRefreshHandler: h.DeviceRefresh,
		DeviceRevokeHandler:  h.DeviceRevoke,
	}
}

// RegisterRoutes registers auth routes on the given Huma API.
func (a API) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:  http.MethodPost,
		Path:    "/auth/register",
		Summary: "Create a new passwordless user",
		Description: "Registers a new user in Logto via the Management API and persists a local ID-only link. " +
			"The user is created without a password and must authenticate via the device flow. " +
			"Returns the Finsplitter user ID and instructions for the next step.",
		Tags: []string{tagAuth},
		Errors: []int{
			http.StatusConflict,
			http.StatusInternalServerError,
		},
	}, a.RegisterHandler)

	huma.Register(api, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/auth/me",
		Summary: "Get the current user's profile",
		Description: "Returns profile information (email, name, username, phone, picture, Finsplitter user ID) " +
			"by reading identity data from the JWT access token claims. " +
			"When no token is provided the response contains empty fields (NeedsSetup defaults to false). " +
			"A newly authenticated user whose Finsplitter record does not yet exist will have NeedsSetup=true " +
			"and must call PATCH /profile/setup to complete registration.",
		Tags:     []string{tagAuth},
		Security: []map[string][]string{{}, {"bearerAuth": {}}},
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusForbidden,
		},
	}, a.MeHandler)

	// Device authorization flow — both endpoints are public (no JWT required).
	huma.Register(api, huma.Operation{
		Method:  http.MethodPost,
		Path:    "/auth/device",
		Summary: "Initiate device authorization flow",
		Description: "Starts the OAuth2 Device Authorization Grant (RFC 8628) flow. " +
			"Accepts an email address and returns a device_code, user_code, and verification_uri. " +
			"The user must visit verification_uri in a browser to approve the request. " +
			"Use POST /auth/device/poll with the device_code to obtain JWT tokens after approval.",
		Tags: []string{tagAuth},
		Errors: []int{
			http.StatusUnprocessableEntity,
			http.StatusInternalServerError,
		},
	}, a.DeviceAuthHandler)

	huma.Register(api, huma.Operation{
		Method:  http.MethodPost,
		Path:    "/auth/device/poll",
		Summary: "Poll for OIDC tokens after user approval",
		Description: "Polls Logto's token endpoint after the user approves the device authorization in the browser. " +
			"Returns JWT access_token, id_token, and refresh_token on success. " +
			"Returns 401 (authorization_pending) while the user has not yet approved, " +
			"400 (device_code_expired) if the code has timed out, and 403 (access_denied) if rejected.",
		Tags: []string{tagAuth},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusUnprocessableEntity,
			http.StatusInternalServerError,
		},
	}, a.DevicePollHandler)

	// Device token refresh — public endpoint (the refresh token is the credential).
	huma.Register(api, huma.Operation{
		Method:  http.MethodPost,
		Path:    "/auth/device/refresh",
		Summary: "Refresh OIDC tokens",
		Description: "Exchanges a refresh token (obtained from POST /auth/device/poll) " +
			"for new access and refresh tokens. Logto rotates refresh tokens, so " +
			"clients must store the returned refresh_token for subsequent refreshes. " +
			"This endpoint is public (no JWT required) — the refresh token is the credential.",
		Tags: []string{tagAuth},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusUnprocessableEntity,
			http.StatusInternalServerError,
		},
	}, a.DeviceRefreshHandler)

	// Device token revocation (RFC 7009) — public endpoint (refresh token is the credential).
	huma.Register(api, huma.Operation{
		Method:  http.MethodPost,
		Path:    "/auth/device/revoke",
		Summary: "Revoke device flow refresh token",
		Description: "Invalidates a refresh token obtained from POST /auth/device/poll. " +
			"The token is immediately unusable for future refreshes. " +
			"This endpoint is public (no JWT required) — the refresh token is the credential.",
		Tags: []string{tagAuth},
		Errors: []int{
			http.StatusUnprocessableEntity,
			http.StatusInternalServerError,
		},
	}, a.DeviceRevokeHandler)
}
