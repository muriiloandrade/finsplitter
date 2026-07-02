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
	meUC       *auth.MeUseCase
}

// NewHandler creates a new auth Handler.
func NewHandler(registerUC *auth.RegisterUseCase, meUC *auth.MeUseCase) *Handler {
	return &Handler{
		registerUC: registerUC,
		meUC:       meUC,
	}
}

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Body struct {
		Name     string `json:"name" required:"true" maxLength:"255" doc:"Display name"`
		Email    string `json:"email" required:"true" maxLength:"255" doc:"Email address"`
		Password string `json:"password" required:"true" minLength:"8" maxLength:"128" doc:"Password (min 8 chars)"`
		Username string `json:"username,omitempty" maxLength:"100" doc:"Desired username (auto-generated from name if omitted)"`
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
		Name:     req.Body.Name,
		Email:    req.Body.Email,
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
		Username   string `json:"username,omitempty" doc:"Username (from Logto)"`
		Email      string `json:"email,omitempty" doc:"Email (from Logto)"`
		Name       string `json:"name,omitempty" doc:"Display name (from Logto)"`
		Phone      string `json:"phone,omitempty" doc:"Phone number (from Logto)"`
		Picture    string `json:"picture,omitempty" doc:"Avatar URL (from Logto)"`
		NeedsSetup bool   `json:"needs_setup" doc:"Whether the user needs to complete profile setup"`
	}
}

// Me returns the current user's auth status and profile info.
// Profile data (username, email) is read from JWT claims, not the local DB.
// GET /auth/me.
func (h *Handler) Me(ctx context.Context, _ *struct{}) (*MeResponse, error) {
	claims := GetUserClaims(ctx)

	if claims == nil {
		// Not authenticated — return empty response (NeedsSetup defaults to false).
		return &MeResponse{}, nil
	}

	output, err := h.meUC.Execute(ctx, auth.MeInput{
		LogtoUserID: claims.Sub,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get user info")
	}

	resp := &MeResponse{}
	resp.Body.ID = output.ID
	resp.Body.Username = claims.Username
	resp.Body.Email = claims.Email
	resp.Body.Name = claims.Name
	resp.Body.Phone = claims.Phone
	resp.Body.Picture = claims.Picture
	resp.Body.NeedsSetup = output.NeedsSetup
	return resp, nil
}
