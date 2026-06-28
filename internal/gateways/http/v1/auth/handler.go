package auth

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
)

// Handler handles auth-related HTTP requests.
type Handler struct {
	registerUC *auth.RegisterUseCase
}

// NewHandler creates a new auth Handler.
func NewHandler(registerUC *auth.RegisterUseCase) *Handler {
	return &Handler{
		registerUC: registerUC,
	}
}

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Body struct {
		Username string `json:"username" required:"true" maxLength:"100" doc:"Desired username"`
		Password string `json:"password" required:"true" minLength:"8" maxLength:"128" doc:"Password (min 8 chars)"`
	}
}

// RegisterResponse is the response for POST /auth/register.
type RegisterResponse struct {
	Body struct {
		RedirectURL string `json:"redirect_url" doc:"URL for the frontend to redirect the user to Logto sign-in"`
		UserID      string `json:"user_id" doc:"The Finsplitter user ID"`
	}
}

// Register handles user registration.
// POST /auth/register.
func (h *Handler) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	output, err := h.registerUC.Execute(ctx, auth.RegisterInput{
		Username: req.Body.Username,
		Password: req.Body.Password,
	})
	if err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			return nil, huma.Error409Conflict("username already taken")
		}
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, huma.Error409Conflict("user already registered")
		}
		return nil, huma.Error500InternalServerError("registration failed")
	}

	resp := &RegisterResponse{}
	resp.Body.RedirectURL = "/auth/sign-in"
	resp.Body.UserID = output.UserID
	return resp, nil
}

// MeResponse is the response for GET /auth/me.
type MeResponse struct {
	Body struct {
		ID         string `json:"id,omitempty" doc:"Finsplitter user ID"`
		Username   string `json:"username,omitempty" doc:"User display name"`
		Email      string `json:"email,omitempty" doc:"User email"`
		NeedsSetup bool   `json:"needs_setup" doc:"Whether the user needs to complete profile setup"`
	}
}

// Me returns the current user's auth status and profile info.
// GET /auth/me.
func (h *Handler) Me(ctx context.Context, _ *struct{}) (*MeResponse, error) {
	claims := GetUserClaims(ctx)

	resp := &MeResponse{}

	if claims == nil {
		// Not authenticated — return empty response (NeedsSetup defaults to false).
		return resp, nil
	}

	resp.Body.Email = claims.Email
	resp.Body.Username = claims.Username
	resp.Body.NeedsSetup = claims.Username == ""

	return resp, nil
}
