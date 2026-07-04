package auth

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// ---------------------------------------------------------------------------
// Request types
// ---------------------------------------------------------------------------

// RequestDeviceAuthRequest is the body for POST /auth/device.
type RequestDeviceAuthRequest struct {
	Body struct {
		Email string `json:"email" required:"true" maxLength:"255" doc:"Email address"`
	}
}

// RequestDeviceAuthResponse is the response for POST /auth/device.
type RequestDeviceAuthResponse struct {
	Body struct {
		DeviceCode              string `json:"device_code" doc:"Device code for polling"`
		UserCode                string `json:"user_code" doc:"Short code to display to user"`
		VerificationURI         string `json:"verification_uri" doc:"URL for user to visit"`
		VerificationURIComplete string `json:"verification_uri_complete" doc:"URL with code pre-filled"`
		ExpiresIn               int    `json:"expires_in" doc:"Lifetime in seconds"`
		Interval                int    `json:"interval" doc:"Polling interval in seconds"`
	}
}

// PollDeviceTokenRequest is the body for POST /auth/device/poll.
type PollDeviceTokenRequest struct {
	Body struct {
		DeviceCode string `json:"device_code" required:"true" doc:"Device code from auth request"`
	}
}

// PollDeviceTokenResponse is the response for POST /auth/device/poll.
type PollDeviceTokenResponse struct {
	Body struct {
		AccessToken  string `json:"access_token" doc:"OIDC access token"`
		IDToken      string `json:"id_token" doc:"OIDC ID token"`
		RefreshToken string `json:"refresh_token,omitempty" doc:"OIDC refresh token"`
		ExpiresIn    int    `json:"expires_in" doc:"Token lifetime in seconds"`
	}
}

// ---------------------------------------------------------------------------
// deviceAuthUseCase interface (satisfied by *auth.RequestDeviceAuthUseCase)
// ---------------------------------------------------------------------------

type deviceAuthUseCase interface {
	Execute(ctx context.Context, input auth.RequestDeviceAuthInput) (*auth.RequestDeviceAuthOutput, error)
}

type devicePollUseCase interface {
	Execute(ctx context.Context, input auth.PollDeviceTokenInput) (*auth.PollDeviceTokenOutput, error)
}

type deviceRefreshUseCase interface {
	Execute(ctx context.Context, input auth.RefreshDeviceTokenInput) (*auth.RefreshDeviceTokenOutput, error)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// DeviceAuth initiates the device authorization flow.
// POST /auth/device.
func (h *Handler) DeviceAuth(ctx context.Context, req *RequestDeviceAuthRequest) (*RequestDeviceAuthResponse, error) {
	output, err := h.deviceAuthUC.Execute(ctx, auth.RequestDeviceAuthInput{
		Email: req.Body.Email,
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidInput) {
			return nil, huma.Error422UnprocessableEntity("email is required")
		}
		return nil, huma.Error500InternalServerError("failed to request device code")
	}

	resp := &RequestDeviceAuthResponse{}
	resp.Body.DeviceCode = output.DeviceCode
	resp.Body.UserCode = output.UserCode
	resp.Body.VerificationURI = output.VerificationURI
	resp.Body.VerificationURIComplete = output.VerificationURIComplete
	resp.Body.ExpiresIn = output.ExpiresIn
	resp.Body.Interval = output.Interval
	return resp, nil
}

// DevicePoll polls for OIDC tokens after the user completes device auth.
// POST /auth/device/poll.
func (h *Handler) DevicePoll(ctx context.Context, req *PollDeviceTokenRequest) (*PollDeviceTokenResponse, error) {
	output, err := h.devicePollUC.Execute(ctx, auth.PollDeviceTokenInput{
		DeviceCode: req.Body.DeviceCode,
	})
	if err != nil {
		if errors.Is(err, logto.ErrDeviceCodePending) {
			return nil, huma.Error401Unauthorized("authorization pending")
		}
		if errors.Is(err, logto.ErrDeviceCodeExpired) {
			return nil, huma.Error400BadRequest("device code expired")
		}
		if errors.Is(err, logto.ErrDeviceCodeAccessDenied) {
			return nil, huma.Error403Forbidden("authorization denied")
		}
		if errors.Is(err, errs.ErrInvalidInput) {
			return nil, huma.Error422UnprocessableEntity("device code is required")
		}
		return nil, huma.Error500InternalServerError("failed to poll device token")
	}

	resp := &PollDeviceTokenResponse{}
	resp.Body.AccessToken = output.AccessToken
	resp.Body.IDToken = output.IDToken
	resp.Body.RefreshToken = output.RefreshToken
	resp.Body.ExpiresIn = output.ExpiresIn
	return resp, nil
}

// ---------------------------------------------------------------------------
// Device Token Refresh
// ---------------------------------------------------------------------------

// DeviceRefreshRequest is the body for POST /auth/device/refresh.
type DeviceRefreshRequest struct {
	Body struct {
		RefreshToken string `json:"refresh_token" required:"true" doc:"The refresh token obtained from device auth"`
	}
}

// DeviceRefreshResponse is the response for POST /auth/device/refresh.
type DeviceRefreshResponse struct {
	Body struct {
		AccessToken  string `json:"access_token" doc:"New JWT access token"`
		IDToken      string `json:"id_token,omitempty" doc:"New ID token"`
		RefreshToken string `json:"refresh_token" doc:"New refresh token (rotated — replace your stored copy)"`
		ExpiresIn    int    `json:"expires_in" doc:"Token lifetime in seconds"`
	}
}

// DeviceRefresh refreshes OIDC tokens using a refresh token.
// The refresh token is obtained from POST /auth/device/poll after user approval.
// POST /auth/device/refresh.
func (h *Handler) DeviceRefresh(ctx context.Context, req *DeviceRefreshRequest) (*DeviceRefreshResponse, error) {
	output, err := h.deviceRefreshUC.Execute(ctx, auth.RefreshDeviceTokenInput{
		RefreshToken: req.Body.RefreshToken,
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidInput) {
			return nil, huma.Error422UnprocessableEntity("refresh_token is required")
		}
		if errors.Is(err, logto.ErrDeviceCodeExpired) {
			return nil, huma.Error400BadRequest("refresh token expired or invalid")
		}
		return nil, huma.Error500InternalServerError("token refresh failed")
	}

	resp := &DeviceRefreshResponse{}
	resp.Body.AccessToken = output.AccessToken
	resp.Body.IDToken = output.IDToken
	resp.Body.RefreshToken = output.RefreshToken
	resp.Body.ExpiresIn = output.ExpiresIn
	return resp, nil
}
