package logto

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/muriiloandrade/finsplitter/pkg/httpclient"
)

// Config holds Logto M2M client configuration.
type Config struct {
	OIDCEndpoint          string
	ManagementBaseURL     string
	ManagementAPIResource string
	ClientID              string
	ClientSecret          string
}

// cachedToken holds an access token with its expiry.
type cachedToken struct {
	accessToken string
	expiresAt   time.Time
}

// Client is a Logto Management API client with automatic token caching.
//
// It uses the shared httpclient under the hood, so retry policy, timeouts,
// transport, and OTel instrumentation are configured at construction via
// httpclient options.
type Client struct {
	httpClient *httpclient.Client
	cfg        Config
	mu         sync.RWMutex
	token      *cachedToken
}

// ClientOption allows callers to override httpclient settings used by the
// Logto client (e.g. transport for testing, timeouts, retry counts).
type ClientOption func(*Client)

// WithLogger sets the resty client's logger to the given *slog.Logger.
// It is a convenience wrapper around WithHTTPClientOptions(httpclient.WithLogger(logger)).
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		httpclient.WithLogger(logger)(c.httpClient)
	}
}

// WithHTTPClientOptions passes extra httpclient options to the underlying
// HTTP client used by the Logto M2M client.
func WithHTTPClientOptions(opts ...httpclient.Option) ClientOption {
	return func(c *Client) {
		for _, opt := range opts {
			opt(c.httpClient)
		}
	}
}

// NewClient creates a new Logto M2M client.
//
// Defaults:
//   - Base URL: cfg.ManagementBaseURL
//   - Retries:  3 (exponential backoff, 500ms–10s)
//   - Timeout:  10s
//   - OTel:     enabled by default (via httpclient.New)
//
// Callers can override these with WithHTTPClientOptions.
const (
	defaultTimeout    = 10 * time.Second
	defaultRetryBase  = 500 * time.Millisecond
	defaultRetryCap   = 5 * time.Second
	tokenExpiryBuffer = 60 // seconds before actual expiry to refresh
)

func NewClient(cfg Config, opts ...ClientOption) *Client {
	defaultOpts := []httpclient.Option{
		httpclient.WithBaseURL(cfg.ManagementBaseURL + "/api"),
		httpclient.WithTimeout(defaultTimeout),
		httpclient.WithRetryWaitTime(defaultRetryBase, defaultRetryCap),
	}

	c := &Client{
		httpClient: httpclient.New(defaultOpts...),
		cfg:        cfg,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ---------------------------------------------------------------------------
// Token management
// ---------------------------------------------------------------------------

// tokenResponse is the OAuth2 token endpoint response.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// getToken returns a valid access token, fetching a new one if the cached
// token is expired (double-checked locking).
func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.token != nil && time.Now().Before(c.token.expiresAt) {
		t := c.token.accessToken
		c.mu.RUnlock()
		return t, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock.
	if c.token != nil && time.Now().Before(c.token.expiresAt) {
		return c.token.accessToken, nil
	}

	resource := c.cfg.ManagementAPIResource
	if resource == "" {
		resource = c.cfg.ManagementBaseURL + "/api"
	}

	formData := map[string]string{
		"grant_type":    "client_credentials",
		"resource":      resource,
		"scope":         "all",
		"client_id":     c.cfg.ClientID,
		"client_secret": c.cfg.ClientSecret,
	}

	var tr tokenResponse
	resp, err := c.httpClient.R(ctx).
		SetFormData(formData).
		SetResult(&tr).
		Post(c.cfg.OIDCEndpoint + "/token")
	if err != nil {
		return "", fmt.Errorf("fetch m2m token: %w", err)
	}
	if resp.StatusCode() >= http.StatusBadRequest {
		return "", fmt.Errorf("%w: status %d", ErrM2MUnauthorized, resp.StatusCode())
	}

	// Cache with a safety buffer before actual expiry.
	// Clamp to zero so sub-60s tokens don't get cached past their lifetime.
	buffer := time.Duration(max(tr.ExpiresIn-tokenExpiryBuffer, 0)) * time.Second

	c.token = &cachedToken{
		accessToken: tr.AccessToken,
		expiresAt:   time.Now().Add(buffer),
	}

	return c.token.accessToken, nil
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// CreateUserRequest is the body for creating a user via Management API.
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"primaryEmail,omitempty"`
}

// CreateUserResponse is the response from creating a user.
type CreateUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// CreateUser creates a new user in Logto via the Management API.
func (c *Client) CreateUser(ctx context.Context, username, password, name, email string) (*CreateUserResponse, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	payload := CreateUserRequest{
		Username: username,
		Password: password,
		Name:     name,
		Email:    email,
	}

	var result CreateUserResponse
	resp, err := c.httpClient.R(ctx).
		SetAuthToken(token).
		SetBody(payload).
		SetResult(&result).
		Post("/users")
	if err != nil {
		return nil, fmt.Errorf("create user request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		return &result, nil
	case http.StatusConflict:
		return nil, ErrUserExists
	case http.StatusUnauthorized:
		return nil, ErrM2MUnauthorized
	default:
		return nil, fmt.Errorf("create user: status %d", resp.StatusCode())
	}
}

// UpdateUserRequest is the body for updating a user via Management API.
// Only non-zero fields are sent — Logto ignores omitted fields.
type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Name     string `json:"name,omitempty"`
	Phone    string `json:"primaryPhone,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// UpdateUser updates a user's profile in Logto via the Management API.
func (c *Client) UpdateUser(ctx context.Context, userID, username, name, phone, picture string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}

	payload := UpdateUserRequest{
		Username: username,
		Name:     name,
		Phone:    phone,
		Avatar:   picture,
	}

	resp, err := c.httpClient.R(ctx).
		SetAuthToken(token).
		SetBody(payload).
		Patch("/users/" + userID)
	if err != nil {
		return fmt.Errorf("update user request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("%w: user %s", ErrUserNotFound, userID)
	case http.StatusUnauthorized:
		return ErrM2MUnauthorized
	default:
		return fmt.Errorf("update user: status %d", resp.StatusCode())
	}
}
