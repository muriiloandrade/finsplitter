package auth_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	openapi "github.com/muriiloandrade/finsplitter/api"
	authUC "github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	authHandler "github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// integrationMockSignIn is a manual mock for the sign-in use case used in
// integration tests.
type integrationMockSignIn struct {
	mock.Mock
}

func (m *integrationMockSignIn) Execute(ctx context.Context, input authUC.SignInInput) (*authUC.SignInOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authUC.SignInOutput), args.Error(1)
}

// newTestSignInHandler sets up a chi router + Huma with the /auth/sign-in
// route wired to a mock use case, returning the router ready for httptest.
func newTestSignInHandler(t *testing.T, mockUC *integrationMockSignIn) *chi.Mux {
	t.Helper()

	r := chi.NewRouter()
	api := humachi.New(r, openapi.NewOpenAPIConfig())

	huma.Register(api, huma.Operation{
		Method: http.MethodPost,
		Path:   "/auth/sign-in",
		Tags:   []string{"Auth"},
	}, func(ctx context.Context, req *authHandler.SignInRequest) (*authHandler.SignInResponse, error) {
		output, err := mockUC.Execute(ctx, authUC.SignInInput{
			Email:    req.Body.Email,
			Password: req.Body.Password,
		})
		if err != nil {
			if err == errs.ErrInvalidCredentials {
				return nil, huma.Error401Unauthorized("invalid email or password")
			}
			return nil, huma.Error500InternalServerError("sign-in failed")
		}

		resp := &authHandler.SignInResponse{}
		resp.Body.AccessToken = output.AccessToken
		resp.Body.IDToken = output.IDToken
		resp.Body.RefreshToken = output.RefreshToken
		resp.Body.ExpiresIn = output.ExpiresIn
		return resp, nil
	})

	return r
}

func TestSignInIntegration_Success(t *testing.T) {
	mockUC := new(integrationMockSignIn)
	mockUC.On("Execute", mock.Anything, authUC.SignInInput{
		Email: "john@example.com", Password: "secret123",
	}).Return(&authUC.SignInOutput{
		AccessToken:  "access_abc",
		IDToken:      "id_def",
		RefreshToken: "refresh_ghi",
		ExpiresIn:    3600,
	}, nil)

	router := newTestSignInHandler(t, mockUC)

	body := `{"email":"john@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "access_abc", resp.AccessToken)
	assert.Equal(t, "id_def", resp.IDToken)
	assert.Equal(t, "refresh_ghi", resp.RefreshToken)
	assert.Equal(t, 3600, resp.ExpiresIn)

	mockUC.AssertExpectations(t)
}

func TestSignInIntegration_InvalidCredentials(t *testing.T) {
	mockUC := new(integrationMockSignIn)
	mockUC.On("Execute", mock.Anything, authUC.SignInInput{
		Email: "wrong@example.com", Password: "badpass",
	}).Return(nil, errs.ErrInvalidCredentials)

	router := newTestSignInHandler(t, mockUC)

	body := `{"email":"wrong@example.com","password":"badpass"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Log raw response for debugging
	rawBody, _ := io.ReadAll(w.Body)
	t.Logf("Status: %d, Body: %s", w.Code, string(rawBody))

	require.Equal(t, http.StatusUnauthorized, w.Code)

	// Huma error responses use a structured error format.
	var errResp struct {
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}
	err := json.Unmarshal(rawBody, &errResp)
	require.NoError(t, err)
	assert.Equal(t, "invalid email or password", errResp.Detail)
	assert.Equal(t, http.StatusUnauthorized, errResp.Status)

	mockUC.AssertExpectations(t)
}

func TestSignInIntegration_ValidationError(t *testing.T) {
	mockUC := new(integrationMockSignIn)
	router := newTestSignInHandler(t, mockUC)

	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Log raw response for debugging
	rawBody, _ := io.ReadAll(w.Body)
	t.Logf("Status: %d, Body: %s", w.Code, string(rawBody))

	// Huma validates required fields — should return 422 Unprocessable Entity.
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var errResp struct {
		Title  string `json:"title"`
		Status int    `json:"status"`
		Errors []struct {
			Message string `json:"message"`
			Field   string `json:"field"`
		} `json:"errors"`
	}
	err := json.Unmarshal(rawBody, &errResp)
	require.NoError(t, err)
	assert.NotEmpty(t, errResp.Title)
	assert.GreaterOrEqual(t, len(errResp.Errors), 1)

	mockUC.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

func TestSignInIntegration_InvalidJSON(t *testing.T) {
	mockUC := new(integrationMockSignIn)
	router := newTestSignInHandler(t, mockUC)

	body := `not-json`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)

	mockUC.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}
