package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	slogctx "github.com/veqryn/slog-context"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	userClaimsKey    contextKey = "user_claims"
	jwksFetchTimeout            = 10 * time.Second
)

// UserClaims holds the JWT claims extracted from the Logto token.
type UserClaims struct {
	Sub      string `json:"sub"`                // Logto user ID
	Username string `json:"username,omitempty"` // May be empty for new users
	Email    string `json:"email,omitempty"`    // User email
}

// GetUserClaims retrieves UserClaims from the context. Returns nil if not present.
func GetUserClaims(ctx context.Context) *UserClaims {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}

// LogtoJWKS holds the JSON Web Key Set from Logto.
type LogtoJWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK is a JSON Web Key.
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// Middleware validates JWTs issued by Logto.
// It extracts the user claims and attaches them to the request context.
type Middleware struct {
	oidcEndpoint string
	appClientID  string
	userRepo     ports.UserRepository
	logger       *slog.Logger

	jwksMu sync.RWMutex
	jwks   *LogtoJWKS
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware(oidcEndpoint, appClientID string, userRepo ports.UserRepository, logger *slog.Logger) *Middleware {
	return &Middleware{
		oidcEndpoint: oidcEndpoint,
		appClientID:  appClientID,
		userRepo:     userRepo,
		logger:       logger,
	}
}

// Protected returns a middleware that requires a valid JWT.
func (m *Middleware) Protected() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := m.logger.With(slog.String("middleware", "auth.protected"))
			ctx := slogctx.NewCtx(r.Context(), logger)

			token := extractBearerToken(r)
			if token == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			claims, err := m.validateToken(ctx, token)
			if err != nil {
				logger.WarnContext(ctx, "Invalid token", slog.Any("error", err))
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Check if user has completed profile setup (has a local user record).
			exists, err := m.userRepo.ExistsByLogtoUserID(ctx, claims.Sub)
			if err != nil {
				logger.ErrorContext(ctx,
					"Failed to check user existence",
					slog.Any("error", err),
				)
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			if !exists {
				writeError(w, http.StatusForbidden, errs.ErrNeedsSetup.Error())
				return
			}

			// Attach claims to context.
			ctx = context.WithValue(ctx, userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Optional returns a middleware that extracts claims if a token is present,
// but allows unauthenticated requests through.
func (m *Middleware) Optional() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token != "" {
				if claims, err := m.validateToken(r.Context(), token); err == nil {
					ctx := context.WithValue(r.Context(), userClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) validateToken(ctx context.Context, tokenString string) (*UserClaims, error) {
	jwks, err := m.fetchJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}

	// Parse the token without verification first to extract the kid header.
	// We need the kid to find the correct JWK for signature verification.
	tokenParser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(m.appClientID),
		jwt.WithIssuer(m.oidcEndpoint),
	)

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("missing kid in JWT header")
		}

		var jwk *JWK
		for i := range jwks.Keys {
			if jwks.Keys[i].Kid == kid {
				jwk = &jwks.Keys[i]
				break
			}
		}
		if jwk == nil {
			return nil, fmt.Errorf("key %q not found in JWKS", kid)
		}

		return jwkToRSAPublicKey(jwk)
	}

	// Custom claims that include optional username and email.
	type customClaims struct {
		jwt.RegisteredClaims

		Username string `json:"username"`
		Email    string `json:"email"`
	}

	var claims customClaims
	parsed, err := tokenParser.ParseWithClaims(tokenString, &claims, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	return &UserClaims{
		Sub:      claims.Subject,
		Username: claims.Username,
		Email:    claims.Email,
	}, nil
}

func (m *Middleware) fetchJWKS(ctx context.Context) (*LogtoJWKS, error) {
	m.jwksMu.RLock()
	if m.jwks != nil {
		jwks := m.jwks
		m.jwksMu.RUnlock()
		return jwks, nil
	}
	m.jwksMu.RUnlock()

	m.jwksMu.Lock()
	defer m.jwksMu.Unlock()

	// Double-check after acquiring write lock.
	if m.jwks != nil {
		return m.jwks, nil
	}

	url := strings.TrimRight(m.oidcEndpoint, "/") + "/jwks"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create jwks request: %w", err)
	}

	client := &http.Client{Timeout: jwksFetchTimeout}
	resp, requestErr := client.Do(req)
	if requestErr != nil {
		return nil, fmt.Errorf("fetch jwks from %s: %w", url, requestErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks endpoint returned status %d", resp.StatusCode)
	}

	var jwks LogtoJWKS
	if decodeErr := json.NewDecoder(resp.Body).Decode(&jwks); decodeErr != nil {
		return nil, fmt.Errorf("decode jwks: %w", decodeErr)
	}

	if len(jwks.Keys) == 0 {
		return nil, errors.New("JWKS contains no keys")
	}

	m.jwks = &jwks
	return &jwks, nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key.
func jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("decode jwk n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("decode jwk e: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(big.NewInt(0).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
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
