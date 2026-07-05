package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/pkg/cache"
	"github.com/muriiloandrade/finsplitter/pkg/httpclient"
	slogctx "github.com/veqryn/slog-context"
)

// jwkFetcher abstracts fetching a JWK Set so the middleware can be tested
// without a real Logto endpoint. Satisfied by *jwkfetch.Client in production.
type jwkFetcher interface {
	Fetch(ctx context.Context, url string) (jwk.Set, error)
}

const jwksCacheKey = "jwks:keyset"

// jwksCacheTTL is how long the JWKS set is cached in Valkey before a refresh.
// This duration must be long enough to amortise the fetch cost but short
// enough to pick up Logto key rotations in reasonable time.
const jwksCacheTTL = 15 * time.Minute

// userInfoHTTPTimeout is the timeout for HTTP calls to Logto's /oidc/me
// endpoint when validating opaque access tokens.
const userInfoHTTPTimeout = 5 * time.Second

//nolint:gochecknoglobals // Path prefix/suffix lists — fixed at init, not mutable state.
var (
	// skipPrefixes are path prefixes that do NOT require authentication.
	// Trailing slashes indicate a directory, so any sub-path is allowed.
	skipPrefixes = []string{
		"/health/",
		"/docs",
		"/openapi",
	}

	// skipExact are exact paths that do NOT require authentication.
	skipExact = []string{
		"/auth/register",
		"/auth/device",
		"/auth/device/poll",
		"/auth/device/refresh",
	}

	// optionalExact are exact paths that do NOT require authentication
	// but will still populate claims from a valid token if one is present.
	optionalExact = []string{
		"/auth/me",
		"/profile/setup",
	}
)

// GetUserClaims retrieves UserClaims from the context. Returns nil if not present.
func GetUserClaims(ctx context.Context) *entity.UserClaims {
	claims, ok := ctx.Value(userClaimsKey).(*entity.UserClaims)
	if !ok {
		return nil
	}
	return claims
}

// WithUserClaims returns a new context with the given UserClaims attached.
func WithUserClaims(ctx context.Context, claims *entity.UserClaims) context.Context {
	return context.WithValue(ctx, userClaimsKey, claims)
}

type contextKey string

const userClaimsKey contextKey = "user_claims"

// Middleware validates JWTs issued by Logto using the jwx library.
// JWKS is fetched on demand and cached in Valkey (via the cache.Client). The
// cache is refreshed every jwksCacheTTL (15 minutes) to handle key rotation.
// When the token is opaque (non-JWT, e.g. device flow), the middleware falls
// back to Logto's /oidc/me endpoint (OIDC UserInfo equivalent) for validation.
type Middleware struct {
	oidcEndpoint string
	appClientID  string
	userRepo     ports.UserRepository
	logger       *slog.Logger

	jwksURL     string
	issuer      string
	userInfoURL string

	jwkClient jwkFetcher // used to fetch JWKS from Logto
	cache     *cache.Client

	httpClient *httpclient.Client
}

// NewMiddleware creates a new auth middleware. The JWKS is not fetched
// until the first authenticated request, so Logto does not need to be
// ready at startup time. The cache client is used to share the JWKS set
// across requests (and potentially across instances).
//
// When the cache client is nil, each request fetches the JWKS from Logto
// directly (no caching).
func NewMiddleware(
	oidcEndpoint, expectedIssuer, appClientID string,
	userRepo ports.UserRepository,
	logger *slog.Logger,
	cacheClient *cache.Client,
) *Middleware {
	return &Middleware{
		oidcEndpoint: oidcEndpoint,
		appClientID:  appClientID,
		userRepo:     userRepo,
		logger:       logger,
		jwksURL:      strings.TrimRight(oidcEndpoint, "/") + "/jwks",
		issuer:       strings.TrimRight(expectedIssuer, "/"),
		userInfoURL:  strings.TrimRight(oidcEndpoint, "/") + "/me",
		jwkClient:    jwkfetch.NewClient(),
		cache:        cacheClient,
		httpClient: httpclient.New(
			httpclient.WithTimeout(userInfoHTTPTimeout),
		),
	}
}

// Claims returns the authenticated user's claims from the context.
// Implements ports.ClaimsProvider.
func (m *Middleware) Claims(ctx context.Context) *entity.UserClaims {
	return GetUserClaims(ctx)
}

// Protected returns a chi middleware that enforces JWT authentication.
// Known public paths are skipped; optional-auth paths populate claims when
// a valid token is present but never reject unauthenticated requests.
func (m *Middleware) Protected() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case isPublicPath(r.URL.Path):
				next.ServeHTTP(w, r)

			case isOptionalPath(r.URL.Path):
				result := m.tryPopulateClaims(r)
				if result.err != nil {
					m.logger.WarnContext(r.Context(), "Optional auth: invalid token",
						slog.Any("error", result.err),
					)
					writeError(w, http.StatusUnauthorized, "invalid or expired token")
					return
				}
				if result.request != nil {
					r = result.request
				}
				next.ServeHTTP(w, r)

			default:
				m.requireAuth(w, r, next)
			}
		})
	}
}

// tryPopulateClaimsResult is returned by tryPopulateClaims to distinguish
// between "no token provided" and "token present but invalid".
type tryPopulateClaimsResult struct {
	request *http.Request // non-nil when claims were successfully attached
	err     error         // non-nil when a token was present but invalid
}

// tryPopulateClaims attaches claims to the request context when a valid
// bearer token is present.
//
// The result encodes three cases:
//   - request != nil, err == nil → token was valid, claims are attached.
//   - request == nil, err == nil → no bearer token was provided.
//   - request == nil, err != nil → a token was provided but is invalid.
func (m *Middleware) tryPopulateClaims(r *http.Request) tryPopulateClaimsResult {
	logger := m.logger.With(slog.String("middleware", "auth.optional"))
	ctx := slogctx.NewCtx(r.Context(), logger)

	token := extractBearerToken(r)
	if token == "" {
		return tryPopulateClaimsResult{}
	}
	claims, err := m.parseAndValidate(ctx, token)
	if err != nil {
		logger.DebugContext(ctx, "Optional auth: invalid token",
			slog.Any("error", err),
		)
		return tryPopulateClaimsResult{err: err}
	}
	ctx = context.WithValue(ctx, userClaimsKey, claims)
	return tryPopulateClaimsResult{request: r.WithContext(ctx)}
}

// requireAuth enforces JWT authentication. On any failure it writes an error
// response and does not call next.
func (m *Middleware) requireAuth(w http.ResponseWriter, r *http.Request, next http.Handler) {
	logger := m.logger.With(slog.String("middleware", "auth.require"))
	ctx := slogctx.NewCtx(r.Context(), logger)

	token := extractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization header")
		return
	}

	claims, err := m.parseAndValidate(ctx, token)
	if err != nil {
		logger.WarnContext(ctx, "Invalid token", slog.Any("error", err))
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	exists, existsErr := m.userRepo.ExistsByLogtoUserID(ctx, claims.Sub)
	if existsErr != nil {
		logger.ErrorContext(ctx,
			"Failed to check user existence",
			slog.Any("error", existsErr),
		)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !exists {
		writeError(w, http.StatusForbidden, errs.ErrNeedsSetup.Error())
		return
	}

	ctx = context.WithValue(ctx, userClaimsKey, claims)
	next.ServeHTTP(w, r.WithContext(ctx))
}

// userInfoResponse maps Logto's /oidc/me response.
type userInfoResponse struct {
	Sub      string `json:"sub"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Picture  string `json:"picture,omitempty"`
}

// parseAndValidate validates the bearer token and returns the user's claims.
// It first attempts JWT parsing + JWKS verification. If the token is opaque
// (non-JWT, e.g. device flow), it falls back to Logto's /oidc/me endpoint.
func (m *Middleware) parseAndValidate(ctx context.Context, tokenString string) (*entity.UserClaims, error) {
	// Try JWT parsing first (faster — no extra HTTP call).
	if tok, err := m.parseWithKeyset(ctx, tokenString); err == nil {
		return m.claimsFromToken(tok)
	}

	// Only fall back to UserInfo if the token is not a JWT (no dots).
	// If it IS a JWT but failed validation, return the error.
	if strings.Contains(tokenString, ".") {
		return nil, errors.New("invalid or expired token")
	}

	return m.parseViaUserInfo(ctx, tokenString)
}

// claimsFromToken extracts UserClaims from a parsed JWT token.
func (m *Middleware) claimsFromToken(tok jwt.Token) (*entity.UserClaims, error) {
	sub, ok := tok.Subject()
	if !ok {
		return nil, errors.New("token missing subject claim")
	}
	return &entity.UserClaims{
		Sub:      sub,
		Username: claimAsString(tok, "username"),
		Email:    claimAsString(tok, "email"),
		Name:     claimAsString(tok, "name"),
		Phone:    claimAsString(tok, "phone"),
		Picture:  claimAsString(tok, "picture"),
	}, nil
}

// userInfoClient returns the HTTP client for UserInfo calls, falling back to
// a one-off client with the standard timeout when the middleware was
// constructed without one (e.g. in unit tests that don't test this path).
func (m *Middleware) userInfoClient() *httpclient.Client {
	if m.httpClient != nil {
		return m.httpClient
	}
	return httpclient.New(httpclient.WithTimeout(userInfoHTTPTimeout))
}

// parseViaUserInfo validates the token by calling Logto's /oidc/me endpoint.
// This is the fallback for opaque access tokens (e.g. device flow).
func (m *Middleware) parseViaUserInfo(ctx context.Context, tokenString string) (*entity.UserClaims, error) {
	if m.userInfoURL == "" {
		return nil, errors.New("userinfo endpoint not configured")
	}

	var ui userInfoResponse
	resp, err := m.userInfoClient().R(ctx).
		SetHeader("Authorization", "Bearer "+tokenString).
		SetResult(&ui).
		Get(m.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("userinfo call: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("userinfo: status %d", resp.StatusCode())
	}

	if ui.Sub == "" {
		return nil, errors.New("userinfo: missing sub claim")
	}

	return &entity.UserClaims{
		Sub:      ui.Sub,
		Username: ui.Username,
		Email:    ui.Email,
		Name:     ui.Name,
		Phone:    ui.Phone,
		Picture:  ui.Picture,
	}, nil
}

// parseWithKeyset fetches (or uses the cached) JWKS, then parses and
// validates the token.
func (m *Middleware) parseWithKeyset(ctx context.Context, tokenString string) (jwt.Token, error) {
	keyset, err := m.getJWKS(ctx)
	if err != nil {
		return nil, err
	}

	tok, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keyset))
	if err != nil {
		return nil, err
	}

	// The audience (aud) claim must match the Logto application's configured
	// API Resource identifier. By default, Logto uses the application's
	// client ID. If a custom API Resource is configured in Logto, set
	// LOGTO_APP_CLIENT_ID to that resource identifier.
	if validateErr := jwt.Validate(tok,
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.appClientID),
	); validateErr != nil {
		return nil, validateErr
	}

	return tok, nil
}

// getJWKS returns the JWKS set, retrieving it from Valkey when possible and
// falling back to a direct Logto fetch on cache miss or cache error.
func (m *Middleware) getJWKS(ctx context.Context) (jwk.Set, error) {
	if m.cache != nil {
		set, err := m.getJWKSFromCache(ctx)
		if err == nil {
			return set, nil
		}
		// Fall through to direct fetch on any cache error (miss, Redis down, ...).
		m.logger.WarnContext(ctx, "JWKS cache miss, fetching from Logto",
			slog.Any("error", err),
		)
	}

	return m.jwkClient.Fetch(ctx, m.jwksURL)
}

// getJWKSFromCache attempts to read the JWKS from Valkey. On a cache hit the
// set is returned immediately. On a miss it fetches from Logto, stores the
// result in Valkey with a TTL, and returns it.
//
// jwk.Set is an interface type, so we use GetBytes + jwk.Parse instead of
// GetJSON (which would fail to unmarshal into a nil interface).
func (m *Middleware) getJWKSFromCache(ctx context.Context) (jwk.Set, error) {
	data, readErr := m.cache.GetBytes(ctx, jwksCacheKey)
	if readErr != nil {
		return nil, fmt.Errorf("cache read: %w", readErr)
	}
	if data != nil {
		set, parseErr := jwk.Parse(data)
		if parseErr != nil {
			return nil, fmt.Errorf("cache parse: %w", parseErr)
		}
		m.logger.DebugContext(ctx, "JWKS cache hit")
		return set, nil
	}

	m.logger.DebugContext(ctx, "JWKS cache miss, fetching from Logto")
	set, fetchErr := m.jwkClient.Fetch(ctx, m.jwksURL)
	if fetchErr != nil {
		return nil, fmt.Errorf("logto fetch: %w", fetchErr)
	}

	if storeErr := m.cache.SetJSON(ctx, jwksCacheKey, set, jwksCacheTTL); storeErr != nil {
		m.logger.WarnContext(ctx, "Failed to store JWKS in cache",
			slog.Any("error", storeErr),
		)
	}

	return set, nil
}

// claimAsString safely extracts a string claim from a jwt.Token.
func claimAsString(tok jwt.Token, key string) string {
	v, ok := tok.Field(key)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// isPublicPath returns true if the path does not require authentication.
func isPublicPath(path string) bool {
	for _, p := range skipPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	for _, p := range skipExact {
		if path == p {
			return true
		}
	}
	return false
}

// isOptionalPath returns true if the path is authentication-optional.
func isOptionalPath(path string) bool {
	for _, p := range optionalExact {
		if path == p {
			return true
		}
	}
	return false
}

// extractBearerToken extracts the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
