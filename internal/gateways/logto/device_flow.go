package logto

import (
	"context"
	"fmt"
	"net/http"
)

// DeviceFlowClient is the interface for initiating and polling the device
// authorization flow. Satisfied by *Client in production.
type DeviceFlowClient interface {
	RequestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error)
	PollDeviceToken(ctx context.Context, deviceCode string) (*DeviceTokenResponse, error)
	RevokeDeviceToken(ctx context.Context, refreshToken string) error
}

var _ DeviceFlowClient = (*Client)(nil)

// DeviceCodeResponse holds the response from Logto's device authorization endpoint.
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DeviceTokenResponse holds the response from polling Logto's token endpoint
// during the device authorization flow.
type DeviceTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`

	// Error fields set on non-success polling responses.
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// RequestDeviceCode initiates the device authorization flow with Logto.
//
// It sends a POST to the OIDC device authorization endpoint using the Native
// App client credentials (DeviceAppClientID). On success it returns a device
// code, user code, and verification URI that the user must visit to complete
// authentication.
func (c *Client) RequestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	if c.cfg.DeviceAppClientID == "" {
		return nil, fmt.Errorf("request device code: %w", ErrAppClientNotConfigured)
	}

	formData := map[string]string{
		fieldClientID: c.cfg.DeviceAppClientID,
		"scope":       "openid profile email offline_access",
	}

	var result DeviceCodeResponse
	resp, err := c.httpClient.R(ctx).
		SetFormData(formData).
		SetResult(&result).
		Post(c.cfg.OIDCEndpoint + "/device/auth")
	if err != nil {
		return nil, fmt.Errorf("request device code: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &result, nil
	default:
		return nil, fmt.Errorf("request device code: status %d", resp.StatusCode())
	}
}

// PollDeviceToken polls Logto's token endpoint to exchange a device code for
// OIDC tokens. The caller should poll at the interval specified in the
// DeviceCodeResponse (minimum 5 seconds between requests).
//
// Returns:
//   - (*DeviceTokenResponse, nil) on success — user has authenticated.
//   - (nil, ErrDeviceCodePending) — user has not yet authenticated.
//   - (nil, ErrDeviceCodeExpired) — device code expired, start a new flow.
//   - (nil, ErrDeviceCodeAccessDenied) — user denied the request.
func (c *Client) PollDeviceToken(ctx context.Context, deviceCode string) (*DeviceTokenResponse, error) {
	if c.cfg.DeviceAppClientID == "" {
		return nil, fmt.Errorf("poll device token: %w", ErrAppClientNotConfigured)
	}

	formData := map[string]string{
		fieldClientID:  c.cfg.DeviceAppClientID,
		fieldGrantType: grantTypeDeviceCode,
		"device_code":  deviceCode,
	}

	// resty v3 uses SetResult for 2xx and SetResultError for non-2xx.
	// Both point to the same struct so result.Error is populated on 4xx too.
	var result DeviceTokenResponse
	resp, err := c.httpClient.R(ctx).
		SetFormData(formData).
		SetResult(&result).
		SetResultError(&result).
		Post(c.cfg.OIDCEndpoint + "/token")
	if err != nil {
		return nil, fmt.Errorf("poll device token: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &result, nil
	case http.StatusBadRequest:
		// Logto returns error details in the response body on bad request.
		switch result.Error {
		case "authorization_pending":
			return nil, ErrDeviceCodePending
		case "slow_down":
			return nil, ErrDeviceCodePending
		case "expired_token":
			return nil, ErrDeviceCodeExpired
		case "access_denied":
			return nil, ErrDeviceCodeAccessDenied
		default:
			return nil, fmt.Errorf("poll device token: %s", result.Error)
		}
	default:
		return nil, fmt.Errorf("poll device token: status %d", resp.StatusCode())
	}
}

// DeviceTokenRefreshResponse maps Logto's /oidc/token response for
// grant_type=refresh_token.
type DeviceTokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`

	// Error fields set on invalid/expired refresh token responses.
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// RefreshDeviceToken exchanges a refresh token for new access and refresh tokens.
//
// Logto rotates refresh tokens, so callers MUST store the returned
// refresh_token for subsequent refreshes.
func (c *Client) RefreshDeviceToken(ctx context.Context, refreshToken string) (*DeviceTokenRefreshResponse, error) {
	if c.cfg.DeviceAppClientID == "" {
		return nil, fmt.Errorf("refresh device token: %w", ErrAppClientNotConfigured)
	}

	formData := map[string]string{
		fieldClientID:     c.cfg.DeviceAppClientID,
		fieldGrantType:    grantTypeRefreshToken,
		fieldRefreshToken: refreshToken,
	}

	var result DeviceTokenRefreshResponse
	resp, err := c.httpClient.R(ctx).
		SetFormData(formData).
		SetResult(&result).
		SetResultError(&result).
		Post(c.cfg.OIDCEndpoint + "/token")
	if err != nil {
		return nil, fmt.Errorf("refresh device token: %w", err)
	}

	if resp.StatusCode() == http.StatusOK {
		return &result, nil
	}

	// Handle error responses (invalid_grant, expired_token, etc.)
	switch result.Error {
	case "invalid_grant":
		return nil, ErrDeviceCodeExpired
	default:
		return nil, fmt.Errorf("refresh device token: %s", result.Error)
	}
}

// RevokeDeviceToken notifies Logto that a refresh token is no longer needed.
// Per RFC 7009, client authentication uses client_id in body (public client).
// Logto accepts form-encoded: client_id + token + token_type_hint=refresh_token.
func (c *Client) RevokeDeviceToken(ctx context.Context, refreshToken string) error {
	if c.cfg.DeviceAppClientID == "" {
		return fmt.Errorf("revoke device token: %w", ErrAppClientNotConfigured)
	}

	formData := map[string]string{
		fieldClientID: c.cfg.DeviceAppClientID,
		"token":       refreshToken,
		// token_type_hint tells the server which kind of token is revoked;
		// for the device flow this is always a refresh token.
		"token_type_hint": fieldRefreshToken,
	}

	var result struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	resp, err := c.httpClient.R(ctx).
		SetFormData(formData).
		SetResult(&result).
		SetResultError(&result).
		Post(c.cfg.OIDCEndpoint + "/token/revocation")
	if err != nil {
		return fmt.Errorf("revoke device token: %w", err)
	}

	if resp.StatusCode() == http.StatusOK {
		return nil
	}

	return fmt.Errorf("revoke device token: %w: %s", ErrDeviceTokenRevoked, result.Error)
}
