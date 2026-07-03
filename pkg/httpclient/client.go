// Package httpclient provides a reusable, configurable HTTP client built on resty v3.
//
// It offers sane defaults with an options pattern so every setting is overridable
// at creation time: retry policy, timeouts, base URL, default headers, transport,
// and middleware hooks.
//
// Basic usage:
//
//	client := httpclient.New(
//	    httpclient.WithBaseURL("https://api.example.com"),
//	    httpclient.WithTimeout(10*time.Second),
//	)
//
//	resp, err := client.R(ctx).
//	    SetHeader("Authorization", "Bearer ...").
//	    SetBody(payload).
//	    Post("/resource")
package httpclient

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"resty.dev/v3"
)

// DefaultRetryCount is the default number of retries.
const DefaultRetryCount = 3

// DefaultRetryWaitTime is the initial wait between retries.
const DefaultRetryWaitTime = 500 * time.Millisecond

// DefaultRetryMaxWaitTime is the upper bound for the exponential backoff.
const DefaultRetryMaxWaitTime = 10 * time.Second

// DefaultTimeout is the default per-request timeout.
const DefaultTimeout = 30 * time.Second

// Client is a reusable HTTP client backed by resty v3.
//
// Zero value is not usable — use New() to construct.
type Client struct {
	resty *resty.Client
}

// Option is a functional option that configures a Client.
type Option func(*Client)

// New builds a Client with sane defaults, then applies the given options.
//
// Defaults:
//   - RetryCount:         3
//   - RetryWaitTime:      500ms (grows exponentially with jitter)
//   - RetryMaxWaitTime:   10s
//   - Timeout:            30s
//   - Retry conditions:   status 429, 5xx, and 0 (network errors)
//   - OpenTelemetry:      enabled (via otelhttp transport)
//
// Callers MUST call Close() when the client is no longer needed to release
// internal resources.
func New(opts ...Option) *Client {
	c := &Client{
		resty: resty.New().
			SetRetryCount(DefaultRetryCount).
			SetRetryWaitTime(DefaultRetryWaitTime).
			SetRetryMaxWaitTime(DefaultRetryMaxWaitTime).
			SetTimeout(DefaultTimeout),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Enable OpenTelemetry instrumentation by default on every client.
	// Applied after user options so a custom transport set via WithTransport
	// is properly wrapped.
	base := c.resty.Transport()
	if base == nil {
		base = http.DefaultTransport
	}
	c.resty.SetTransport(otelhttp.NewTransport(base))

	return c
}

// Close releases internal resources held by the resty client.
func (c *Client) Close() {
	_ = c.resty.Close()
}

// ---------------------------------------------------------------------------
// Request factory
// ---------------------------------------------------------------------------

// R returns a fresh resty.Request wired with the supplied context.
//
// The ctx controls cancellation and deadline propagation for HTTP calls made
// with the returned request.
func (c *Client) R(ctx context.Context) *resty.Request {
	return c.resty.R().SetContext(ctx)
}

// ---------------------------------------------------------------------------
// Underlying client access
// ---------------------------------------------------------------------------

// Resty exposes the underlying resty client for advanced configuration that is
// not covered by the Option helpers.
func (c *Client) Resty() *resty.Client { return c.resty }

// ---------------------------------------------------------------------------
// Logger adapter — bridges resty's Logger interface to log/slog
// ---------------------------------------------------------------------------

// restyLogger adapts a *slog.Logger to resty.dev/v3's Logger interface.
type restyLogger struct {
	inner *slog.Logger
}

func (l *restyLogger) Errorf(format string, v ...any) {
	l.inner.Error(format, "args", v)
}

func (l *restyLogger) Warnf(format string, v ...any) {
	l.inner.Warn(format, "args", v)
}

func (l *restyLogger) Debugf(format string, v ...any) {
	l.inner.Debug(format, "args", v)
}

// ---------------------------------------------------------------------------
// Functional options
// ---------------------------------------------------------------------------

// WithRetryCount overrides the maximum number of retry attempts.
//
// A value of 0 disables retries entirely.
func WithRetryCount(n int) Option {
	return func(c *Client) {
		c.resty.SetRetryCount(n)
	}
}

// WithRetryWaitTime sets the initial (min) and maximum delay between retries.
// The actual wait grows exponentially with jitter between these bounds.
func WithRetryWaitTime(waitMin, waitMax time.Duration) Option {
	return func(c *Client) {
		c.resty.SetRetryWaitTime(waitMin)
		c.resty.SetRetryMaxWaitTime(waitMax)
	}
}

// WithTimeout sets the per-request timeout (dial, TLS handshake, request,
// response body).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.resty.SetTimeout(d)
	}
}

// WithBaseURL sets a base URL that is prepended to every request path.
//
//	client := httpclient.New(httpclient.WithBaseURL("https://api.dev"))
//	resp, err := client.R(ctx).Get("/v1/users")   // → GET https://api.dev/v1/users
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.resty.SetBaseURL(url)
	}
}

// WithHeaders sets default headers that are sent on every request.
// Existing per-request headers take precedence.
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		c.resty.SetHeaders(headers)
	}
}

// WithHeader is a convenience option that sets a single default header.
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.resty.SetHeader(key, value)
	}
}

// WithRetryCondition adds a custom retry condition to the built-in defaults.
// The condition is evaluated after the default conditions — retries happen
// when ANY condition returns true.
func WithRetryCondition(condition resty.RetryConditionFunc) Option {
	return func(c *Client) {
		c.resty.AddRetryConditions(condition)
	}
}

// WithRetryHook registers a callback that is invoked after each failed attempt.
func WithRetryHook(hook resty.RetryHookFunc) Option {
	return func(c *Client) {
		c.resty.AddRetryHooks(hook)
	}
}

// WithTransport replaces the underlying http.RoundTripper (useful for testing
// with httptest or injecting custom TLS config).
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) {
		c.resty.SetTransport(transport)
	}
}

// WithOpenTelemetry wraps the client's http.RoundTripper with
// [otelhttp.NewTransport], adding distributed tracing, metrics, and context
// propagation to every outgoing request made through this client.
//
// The tracer and meter providers are obtained from the global OTel SDK
// (otel.GetTracerProvider / otel.GetMeterProvider). Callers that need custom
// providers should use WithTransport instead with an otelhttp Transport
// configured manually.
func WithOpenTelemetry() Option {
	return func(c *Client) {
		base := c.resty.Transport()
		if base == nil {
			base = http.DefaultTransport
		}
		c.resty.SetTransport(otelhttp.NewTransport(base))
	}
}

// WithRequestMiddleware appends a request middleware that is invoked before
// every HTTP call during request preparation.
func WithRequestMiddleware(m resty.RequestMiddleware) Option {
	return func(c *Client) {
		c.resty.AddRequestMiddleware(m)
	}
}

// WithResponseMiddleware appends a response middleware that is invoked after
// every HTTP call (including failed retries).
func WithResponseMiddleware(m resty.ResponseMiddleware) Option {
	return func(c *Client) {
		c.resty.AddResponseMiddleware(m)
	}
}

// WithScheme sets the URL scheme (e.g. "http", "https").
func WithScheme(scheme string) Option {
	return func(c *Client) {
		c.resty.SetScheme(scheme)
	}
}

// WithPathParam sets a default path parameter (replaces {param} in URL paths).
func WithPathParam(name, value string) Option {
	return func(c *Client) {
		c.resty.SetPathParam(name, value)
	}
}

// WithQueryParam sets a default query parameter sent on every request.
func WithQueryParam(name, value string) Option {
	return func(c *Client) {
		c.resty.SetQueryParam(name, value)
	}
}

// WithCookies sets default cookies sent on every request.
func WithCookies(cookies ...*http.Cookie) Option {
	return func(c *Client) {
		c.resty.SetCookies(cookies)
	}
}

// WithBasicAuth sets the default HTTP Basic Authentication credentials.
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		c.resty.SetBasicAuth(username, password)
	}
}

// WithAuthToken sets the default Authorization Bearer token.
func WithAuthToken(token string) Option {
	return func(c *Client) {
		c.resty.SetAuthToken(token)
	}
}

// WithResponseSaveDirectory sets the directory where response bodies are saved
// when the request has a SetOutputFileName or when SetResponseSaveToFile(true)
// is enabled globally.
func WithResponseSaveDirectory(dir string) Option {
	return func(c *Client) {
		c.resty.SetResponseSaveDirectory(dir)
	}
}

// WithResponseSaveToFile enables or disables automatic saving of response
// bodies to files in the configured save directory.
func WithResponseSaveToFile(save bool) Option {
	return func(c *Client) {
		c.resty.SetResponseSaveToFile(save)
	}
}

// WithMethodGetAllowPayload configures whether GET requests are allowed to
// carry a body (non-standard but required by some APIs).
func WithMethodGetAllowPayload(allow bool) Option {
	return func(c *Client) {
		c.resty.SetMethodGetAllowPayload(allow)
	}
}

// WithMethodDeleteAllowPayload configures whether DELETE requests are allowed
// to carry a payload.
func WithMethodDeleteAllowPayload(allow bool) Option {
	return func(c *Client) {
		c.resty.SetMethodDeleteAllowPayload(allow)
	}
}

// WithLogger sets the resty client's logger to the given *slog.Logger.
// Resty will use it for debug, warning, and error messages during request
// execution, retries, and middleware processing.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.resty.SetLogger(&restyLogger{inner: logger})
	}
}
