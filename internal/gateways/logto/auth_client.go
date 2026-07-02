package logto

import (
	"context"
	"fmt"
	"net/http"
)

// LogtoTokenResponse holds the response from Logto's OIDC token endpoint
// when authenticating a user via the Resource Owner Password Credentials grant.
type LogtoTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// AuthenticateUser authenticates a user against Logto's OIDC token endpoint
// using the Resource Owner Password Credentials (ROPC) grant.
//
// The method uses the application's OIDC client credentials (AppClientID /
// AppClientSecret) — not the Management API credentials — and calls the
// OIDC-compliant /token endpoint directly.
//
// On success it returns access, ID, and optionally refresh tokens that the
// client can use to authenticate subsequent requests to Finsplitter.
func (c *Client) AuthenticateUser(
	ctx context.Context, email, password string,
) (*LogtoTokenResponse, error) {
	if c.cfg.AppClientID == "" {
		return nil, fmt.Errorf("authenticate user: %w", ErrAppClientNotConfigured)
	}

	formData := map[string]string{
		"grant_type":    "password",
		"username":      email,
		"password":      password,
		"scope":         "openid profile email",
		"client_id":     c.cfg.AppClientID,
		"client_secret": c.cfg.AppClientSecret,
	}

	var result LogtoTokenResponse
	resp, err := c.httpClient.R(ctx).
		SetFormData(formData).
		SetResult(&result).
		Post(c.cfg.OIDCEndpoint + "/token")
	if err != nil {
		return nil, fmt.Errorf("authenticate user request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return &result, nil
	case http.StatusBadRequest, http.StatusUnauthorized:
		return nil, ErrInvalidCredentials
	default:
		return nil, fmt.Errorf("authenticate user: status %d", resp.StatusCode())
	}
}
