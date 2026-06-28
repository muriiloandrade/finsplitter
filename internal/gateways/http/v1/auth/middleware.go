package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	slogctx "github.com/veqryn/slog-context"
)

const (
	jwksRefreshInterval = 15 * time.Minute
	jwksFetchTimeout    = 10 * time.Second
)

// enableJWKSRefreshOnMiss controls whether the middleware re-fetches the JWKS
// when no matching key is found. This handles Logto key rotation without a
// server restart.
const enableJWKSRefreshOnMiss = true

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
// JWKS is fetched and cached via jwkfetch.Cache, which auto-refreshes
// in the background to handle key rotation.
type Middleware struct {
	oidcEndpoint string
	appClientID  string
	userRepo     ports.UserRepository
	logger       *slog.Logger

	jwksURL  string
	issuer   string
	jwkCache *jwkfetch.Cache
}

// NewMiddleware creates a new auth middleware. It initialises a background-
// refreshed JWKS cache using the Logto OIDC endpoint.
func NewMiddleware(oidcEndpoint, appClientID string, userRepo ports.UserRepository, logger *slog.Logger) *Middleware {
	base := strings.TrimRight(oidcEndpoint, "/")
	jwksURL := base + "/jwks"
	issuer := base

	// Use background context for the cache lifecycle so the background
	// refresh goroutine lives for the entire app lifetime.
	cache, err := jwkfetch.NewCache(context.Background(), httprc.NewClient())
	if err != nil {
		// Should never happen with the default client.
		logger.Warn("Failed to create JWKS cache, falling back to one-shot fetch",
			slog.Any("error", err),
		)
		return &Middleware{
			oidcEndpoint: oidcEndpoint,
			appClientID:  appClientID,
			userRepo:     userRepo,
			logger:       logger,
			jwksURL:      jwksURL,
			issuer:       issuer,
		}
	}

	// Register the JWKS URL in a background goroutine so it doesn't block
	// app startup. Once Logto is ready the initial fetch succeeds, and the
	// cache auto-refreshes from then on. Until registration completes,
	// requests fall back to one-shot fetches.
	go func() {
		regCtx, regCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer regCancel()
		if regErr := cache.Register(regCtx, jwksURL,
			jwkfetch.WithMinInterval(jwksRefreshInterval),
		); regErr != nil {
			logger.Warn("Failed to register JWKS URL in cache",
				slog.Any("error", regErr),
			)
		}
	}()

	return &Middleware{
		oidcEndpoint: oidcEndpoint,
		appClientID:  appClientID,
		userRepo:     userRepo,
		logger:       logger,
		jwksURL:      jwksURL,
		issuer:       issuer,
		jwkCache:     cache,
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

// parseAndValidate parses a JWT token string, verifies its signature against
// the cached JWKS, and validates standard claims (issuer, audience).
// It handles key rotation by re-fetching the JWKS once when no key matches.
func (m *Middleware) parseAndValidate(ctx context.Context, tokenString string) (*UserClaims, error) {
	tok, err := m.parseWithCache(ctx, tokenString)
	if err != nil {
		if enableJWKSRefreshOnMiss && m.jwkCache != nil {
			// Force a refresh of the JWKS cache — the key may have been
			// rotated. Retry exactly once.
			if _, refreshErr := m.jwkCache.Refresh(ctx, m.jwksURL); refreshErr != nil {
				m.logger.WarnContext(ctx, "Failed to refresh JWKS cache after validation miss",
					slog.Any("error", refreshErr),
				)
			}
			tok, err = m.parseWithCache(ctx, tokenString)
		}
	}
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

// lookupJWKS returns a jwk.Set from the cache, or nil if the cache is nil.
func (m *Middleware) lookupJWKS(ctx context.Context) (jwk.Set, error) {
	if m.jwkCache == nil {
		return nil, errors.New("jwks cache not available")
	}
	return m.jwkCache.Lookup(ctx, m.jwksURL)
}

// parseWithCache attempts to parse and validate the token using the cached
// JWKS. When the cache is nil or a lookup fails, it falls back to a one-shot
// fetch (e.g. when Logto wasn't ready during the initial registration).
func (m *Middleware) parseWithCache(ctx context.Context, tokenString string) (jwt.Token, error) {
	keyset, err := m.lookupJWKS(ctx)
	if err != nil {
		// Cache unavailable — fall back to one-shot fetch.
		client := jwkfetch.NewClient()
		keyset, err = client.Fetch(ctx, m.jwksURL)
		if err != nil {
			return nil, err
		}
	}

	tok, err := jwt.Parse([]byte(tokenString), jwt.WithKeySet(keyset))
	if err != nil {
		return nil, err
	}

	// Validate standard claims.
	if validateErr := jwt.Validate(tok,
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.appClientID),
	); validateErr != nil {
		return nil, validateErr
	}

	return tok, nil
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
// Directory-like prefixes (with trailing /) match any sub-path; individual
// paths use exact matching to avoid unintended matches.
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
// Uses exact matching to avoid unintended matches.
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
