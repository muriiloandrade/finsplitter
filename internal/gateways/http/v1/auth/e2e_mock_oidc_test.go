//go:build e2e

package auth_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

// deviceCodeState tracks the lifecycle of a device authorization code.
type deviceCodeState string

const (
	deviceCodePending  deviceCodeState = "pending"
	deviceCodeExpired  deviceCodeState = "expired"
	deviceCodeDenied   deviceCodeState = "denied"
	deviceCodeApproved deviceCodeState = "approved"
)

// deviceCode represents a device authorization code in the mock provider.
type deviceCode struct {
	code      string
	state     deviceCodeState
	userCode  string
	userEmail string // set when the code is approved
	expiresAt time.Time
}

// oidcUser represents a user stored in the mock OIDC provider's in-memory store.
type oidcUser struct {
	ID       string
	Email    string
	Password string
	Name     string
	Username string
}

const mockOIDCKeyID = "mock-oidc-key-1"

// mockOIDCProvider simulates Logto's OIDC + Management API endpoints
// using an in-memory user store and ephemeral EC P-256 keys for JWT signing.
//
// The provider handles:
//   - POST /oidc/token — ROPC grant (grant_type=password)
//   - POST /oidc/token — client_credentials grant (for M2M)
//   - GET  /oidc/jwks  — returns the public JWK Set
//   - POST /api/users  — creates a user (Management API)
//
// JWTs are signed with a real EC P-256 key so Finsplitter's middleware
// validates them for real (signature verification, issuer/audience checks).
//
// To support jwx's jwk.WithKeySet (which requires kid matching by default),
// the signing key carries a stable kid ("mock-oidc-key-1") that is included
// in both the JWT header and the JWKS response.
type mockOIDCProvider struct {
	server             *http.Server
	listener           net.Listener
	users              map[string]*oidcUser   // keyed by email
	deviceCodes        map[string]*deviceCode // keyed by device_code
	refreshTokens      map[string]*oidcUser   // keyed by refresh_token
	deviceAuthClientID string                 // expected client_id for device auth
	privKey            *ecdsa.PrivateKey      // raw key for signing JWTs
	signingJWK         jwk.Key                // private JWK with kid (for signing, keeps kid in header)
	publicJWK          jwk.Key                // public JWK with kid (served via JWKS endpoint)
	issuer             string
	audience           string
	mu                 sync.RWMutex
}

// newMockOIDCProvider creates a new mock OIDC provider, starts the HTTP server,
// and returns it. Callers must call Close() when done.
func newMockOIDCProvider(audience string) (*mockOIDCProvider, error) {
	rawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate oidc key: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("mock oidc listen: %w", err)
	}

	baseURL := fmt.Sprintf("http://%s", listener.Addr().String())

	// Import the private key to a jwk.Key so we can set kid for kid-based
	// key matching in jwx's jwt.WithKeySet.
	signingJWK, err := jwk.Import[jwk.Key](rawKey)
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("import signing jwk: %w", err)
	}
	if err := signingJWK.Set(jwk.KeyIDKey, mockOIDCKeyID); err != nil {
		listener.Close()
		return nil, fmt.Errorf("set kid on signing key: %w", err)
	}

	// Derive the public JWK (same kid).
	pubJWK, err := jwk.PublicKeyOf(signingJWK)
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("public key of signing jwk: %w", err)
	}
	// Set alg on the public JWK. jwx v4's jwt.WithKeySet needs the algorithm
	// parameter on the key itself; it does not infer it from the JWT header's
	// "alg" field like jwt.WithKey does.
	if err := pubJWK.Set(jwk.AlgorithmKey, jwa.ES256()); err != nil {
		listener.Close()
		return nil, fmt.Errorf("set alg on public key: %w", err)
	}

	m := &mockOIDCProvider{
		users:              make(map[string]*oidcUser),
		deviceCodes:        make(map[string]*deviceCode),
		refreshTokens:      make(map[string]*oidcUser),
		deviceAuthClientID: "test-device-client-id",
		signingJWK:         signingJWK,
		publicJWK:          pubJWK,
		issuer:             baseURL + "/oidc",
		audience:           audience,
		listener:           listener,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/oidc/device/auth", m.handleDeviceAuth)
	mux.HandleFunc("/oidc/token", m.handleToken)
	mux.HandleFunc("/oidc/jwks", m.handleJWKS)
	mux.HandleFunc("/api/users", m.handleCreateUser)

	m.server = &http.Server{Handler: mux}
	go func() { _ = m.server.Serve(listener) }()

	return m, nil
}

// Close shuts down the mock OIDC server.
func (m *mockOIDCProvider) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = m.server.Shutdown(ctx)
}

// URL returns the base URL of the mock OIDC server (e.g. http://127.0.0.1:54321).
func (m *mockOIDCProvider) URL() string {
	return fmt.Sprintf("http://%s", m.listener.Addr().String())
}

// OIDCEndpoint returns the OIDC endpoint URL (with /oidc prefix).
func (m *mockOIDCProvider) OIDCEndpoint() string {
	return m.issuer
}

// ManagementBaseURL returns the base URL for Management API calls.
func (m *mockOIDCProvider) ManagementBaseURL() string {
	return m.URL()
}

// SetDeviceAuthClientID overrides the default client ID for device auth.
func (m *mockOIDCProvider) SetDeviceAuthClientID(id string) {
	m.deviceAuthClientID = id
}

// NewDeviceCode creates a new device code in the pending state and returns
// the device code and user code.
func (m *mockOIDCProvider) NewDeviceCode() (string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	dc := uuid.Must(uuid.NewV4()).String()
	uc := uuid.Must(uuid.NewV4()).String()[:8]
	m.deviceCodes[dc] = &deviceCode{
		code:      dc,
		state:     deviceCodePending,
		userCode:  uc,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	return dc, uc
}

// ApproveDeviceCode marks a device code as approved by a user. Returns false
// if the device code does not exist or is not in pending state.
func (m *mockOIDCProvider) ApproveDeviceCode(code, userEmail string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	dc, ok := m.deviceCodes[code]
	if !ok || dc.state != deviceCodePending {
		return false
	}
	dc.state = deviceCodeApproved
	dc.userEmail = userEmail
	return true
}

// ExpireDeviceCode marks a device code as expired.
func (m *mockOIDCProvider) ExpireDeviceCode(deviceCode string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	dc, ok := m.deviceCodes[deviceCode]
	if !ok {
		return false
	}
	dc.state = deviceCodeExpired
	return true
}

// DenyDeviceCode marks a device code as denied by user.
func (m *mockOIDCProvider) DenyDeviceCode(deviceCode string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	dc, ok := m.deviceCodes[deviceCode]
	if !ok {
		return false
	}
	dc.state = deviceCodeDenied
	return true
}

// AddUser adds a user to the mock's in-memory store and returns the
// generated Logto user ID.
func (m *mockOIDCProvider) AddUser(email, password, name, username string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.Must(uuid.NewV4()).String()
	m.users[email] = &oidcUser{
		ID:       id,
		Email:    email,
		Password: password,
		Name:     name,
		Username: username,
	}
	return id
}

// newRefreshToken generates a unique refresh token and stores it for the user.
// Returns the raw token string. The caller must hold m.mu.
func (m *mockOIDCProvider) newRefreshToken(user *oidcUser) string {
	rt := uuid.Must(uuid.NewV4()).String()
	m.refreshTokens[rt] = user
	return rt
}

// createToken signs a JWT with the mock's private key containing the
// standard OIDC claims expected by Finsplitter's auth middleware.
func (m *mockOIDCProvider) createToken(user *oidcUser) (string, error) {
	now := time.Now()
	tok := jwt.New()

	if err := tok.Set("sub", user.ID); err != nil {
		return "", fmt.Errorf("set sub: %w", err)
	}
	if err := tok.Set("email", user.Email); err != nil {
		return "", fmt.Errorf("set email: %w", err)
	}
	if err := tok.Set("username", user.Username); err != nil {
		return "", fmt.Errorf("set username: %w", err)
	}
	if err := tok.Set("name", user.Name); err != nil {
		return "", fmt.Errorf("set name: %w", err)
	}
	if err := tok.Set("iss", m.issuer); err != nil {
		return "", fmt.Errorf("set iss: %w", err)
	}
	if err := tok.Set("aud", m.audience); err != nil {
		return "", fmt.Errorf("set aud: %w", err)
	}
	if err := tok.Set("exp", now.Add(1*time.Hour).Unix()); err != nil {
		return "", fmt.Errorf("set exp: %w", err)
	}
	if err := tok.Set("iat", now.Unix()); err != nil {
		return "", fmt.Errorf("set iat: %w", err)
	}

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.ES256(), m.signingJWK))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return string(signed), nil
}

// --- HTTP handlers ---------------------------------------------------------

func (m *mockOIDCProvider) handleDeviceAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	clientID := r.FormValue("client_id")
	if clientID != m.deviceAuthClientID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid client"})
		return
	}

	dc, uc := m.NewDeviceCode()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"device_code":               dc,
		"user_code":                 uc,
		"verification_uri":          fmt.Sprintf("%s/device", m.URL()),
		"verification_uri_complete": fmt.Sprintf("%s/device?user_code=%s", m.URL(), uc),
		"expires_in":                300,
		"interval":                  5,
	})
}

func (m *mockOIDCProvider) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	switch grantType {
	case "password":
		m.handleROPCToken(w, r)
	case "client_credentials":
		m.handleM2MToken(w, r)
	case "urn:ietf:params:oauth:grant-type:device_code":
		m.handleDeviceCodeToken(w, r)
	case "refresh_token":
		m.handleRefreshTokenGrant(w, r)
	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

func (m *mockOIDCProvider) handleROPCToken(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("username")
	password := r.FormValue("password")

	m.mu.RLock()
	user, ok := m.users[email]
	m.mu.RUnlock()

	if !ok || user.Password != password {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_grant"})
		return
	}

	token, err := m.createToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token creation failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  token,
		"id_token":      token,
		"refresh_token": "mock_refresh_token",
		"expires_in":    3600,
		"token_type":    "Bearer",
	})
}

func (m *mockOIDCProvider) handleM2MToken(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": "mock_m2m_access_token",
		"expires_in":   3600,
		"token_type":   "Bearer",
	})
}

func (m *mockOIDCProvider) handleDeviceCodeToken(w http.ResponseWriter, r *http.Request) {
	deviceCode := r.FormValue("device_code")
	if deviceCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	m.mu.RLock()
	dc, ok := m.deviceCodes[deviceCode]
	m.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid_grant",
		})
		return
	}

	if time.Now().After(dc.expiresAt) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "expired_token",
		})
		return
	}

	switch dc.state {
	case deviceCodePending:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "authorization_pending",
		})
	case deviceCodeDenied:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "access_denied",
		})
	case deviceCodeExpired:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "expired_token",
		})
	case deviceCodeApproved:
		// Look up the user by email and issue a token.
		m.mu.RLock()
		user, ok := m.users[dc.userEmail]
		m.mu.RUnlock()

		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "user_not_found",
			})
			return
		}

		token, err := m.createToken(user)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token creation failed"})
			return
		}

		rt := m.newRefreshToken(user)

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"access_token":  token,
			"id_token":      token,
			"refresh_token": rt,
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid_grant",
		})
	}
}

func (m *mockOIDCProvider) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	user, ok := m.refreshTokens[refreshToken]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid_grant",
		})
		return
	}

	token, err := m.createToken(user)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token creation failed"})
		return
	}

	// Rotate refresh token.
	delete(m.refreshTokens, refreshToken)
	newRT := m.newRefreshToken(user)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  token,
		"id_token":      token,
		"refresh_token": newRT,
		"expires_in":    3600,
		"token_type":    "Bearer",
	})
}

func (m *mockOIDCProvider) handleJWKS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyJSON, err := json.Marshal(m.publicJWK)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "jwk marshal failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"keys": []json.RawMessage{keyJSON},
	})
}

func (m *mockOIDCProvider) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Email    string `json:"primaryEmail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	userID := m.AddUser(body.Email, body.Password, body.Name, body.Username)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":       userID,
		"username": body.Username,
	})
}

// --- helpers ---------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
