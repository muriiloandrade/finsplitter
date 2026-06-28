package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	slogctx "github.com/veqryn/slog-context"
)

const jwksRefreshInterval = 15 * time.Minute

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
	}

	// optionalExact are exact paths that do NOT require authentication
	// but will still populate claims from a valid token if one is present.
	optionalExact = []string{
		"/auth/me",
	}
)

// UserClaims holds the JWT claims extracted from the Logto token.
type UserClaims struct {
	Sub      string `json:"sub"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}

// GetUserClaims retrieves UserClaims from the context. Returns nil if not present.
func GetUserClaims(ctx context.Context) *UserClaims {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}

type contextKey string

const userClaimsKey contextKey = "user_claims"

// Middleware validates JWTs issued by Logto using the jwx library.
// JWKS is fetched on demand and cached in-memory. The cache is
// refreshed every jwksRefreshInterval to handle key rotation.
type Middleware struct {
	oidcEndpoint string
	appClientID  string
	userRepo     ports.UserRepository
	logger       *slog.Logger

	jwksURL string
	issuer  string
	client  *jwkfetch.Client

	jwkSetMu   sync.RWMutex
	jwkSet     jwk.Set   // cached JWKS, nil until first fetch
	jwkSetTime time.Time // when jwkSet was last fetched
}

// NewMiddleware creates a new auth middleware. The JWKS is not fetched
// until the first authenticated request, so Logto does not need to be
// ready at startup time.
func NewMiddleware(oidcEndpoint, appClientID string, userRepo ports.UserRepository, logger *slog.Logger) *Middleware {
	base := strings.TrimRight(oidcEndpoint, "/")
	return &Middleware{
		oidcEndpoint: oidcEndpoint,
		appClientID:  appClientID,
		userRepo:     userRepo,
		logger:       logger,
		jwksURL:      base + "/jwks",
		issuer:       base,
		client:       jwkfetch.NewClient(),
	}
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
				if modified := m.tryPopulateClaims(r); modified != nil {
					r = modified
				}
				next.ServeHTTP(w, r)

			default:
				m.requireAuth(w, r, next)
			}
		})
	}
}

// tryPopulateClaims attaches claims to the request context when a valid
// bearer token is present. Never rejects the request.
func (m *Middleware) tryPopulateClaims(r *http.Request) *http.Request {
	logger := m.logger.With(slog.String("middleware", "auth.optional"))
	ctx := slogctx.NewCtx(r.Context(), logger)

	token := extractBearerToken(r)
	if token == "" {
		return nil
	}
	claims, err := m.parseAndValidate(ctx, token)
	if err != nil {
		logger.DebugContext(ctx, "Optional auth: invalid token",
			slog.Any("error", err),
		)
		return nil
	}
	ctx = context.WithValue(ctx, userClaimsKey, claims)
	return r.WithContext(ctx)
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

// parseAndValidate parses a JWT, verifies its signature against the cached
// JWKS (refreshing if stale), and validates standard claims.
func (m *Middleware) parseAndValidate(ctx context.Context, tokenString string) (*UserClaims, error) {
	tok, err := m.parseWithKeyset(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	sub, ok := tok.Subject()
	if !ok {
		return nil, errors.New("token missing subject claim")
	}

	return &UserClaims{
		Sub:      sub,
		Username: claimAsString(tok, "username"),
		Email:    claimAsString(tok, "email"),
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

	if validateErr := jwt.Validate(tok,
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.appClientID),
	); validateErr != nil {
		return nil, validateErr
	}

	return tok, nil
}

// getJWKS returns the cached JWKS, fetching it from Logto on first call or
// when the cached set is stale (older than jwksRefreshInterval). This TTL-
// based approach handles key rotation without error-based cache invalidation,
// so bad-token requests never trigger unnecessary re-fetches.
func (m *Middleware) getJWKS(ctx context.Context) (jwk.Set, error) {
	// Fast path: already cached and fresh.
	m.jwkSetMu.RLock()
	if m.jwkSet != nil && time.Since(m.jwkSetTime) < jwksRefreshInterval {
		set := m.jwkSet
		m.jwkSetMu.RUnlock()
		return set, nil
	}
	m.jwkSetMu.RUnlock()

	// Slow path: fetch and cache.
	m.jwkSetMu.Lock()
	defer m.jwkSetMu.Unlock()

	// Double-check after acquiring write lock.
	if m.jwkSet != nil && time.Since(m.jwkSetTime) < jwksRefreshInterval {
		return m.jwkSet, nil
	}

	set, err := m.client.Fetch(ctx, m.jwksURL)
	if err != nil {
		return nil, err
	}
	m.jwkSet = set
	m.jwkSetTime = time.Now()
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

// Optional returns a chi middleware that extracts claims if a token is
// present but allows unauthenticated requests through.
func (m *Middleware) Optional() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token != "" {
				if claims, err := m.parseAndValidate(r.Context(), token); err == nil {
					ctx := context.WithValue(r.Context(), userClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
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
