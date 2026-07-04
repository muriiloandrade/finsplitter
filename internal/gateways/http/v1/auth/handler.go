package auth

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

// Handler handles auth-related HTTP requests.
type Handler struct {
	registerUC   *auth.RegisterUseCase
	meUC         *auth.MeUseCase
	deviceAuthUC deviceAuthUseCase
	devicePollUC devicePollUseCase
}

// NewHandler creates a new auth Handler.
func NewHandler(
	registerUC *auth.RegisterUseCase,
	meUC *auth.MeUseCase,
	deviceAuthUC deviceAuthUseCase,
	devicePollUC devicePollUseCase,
) *Handler {
	return &Handler{
		registerUC:   registerUC,
		meUC:         meUC,
		deviceAuthUC: deviceAuthUC,
		devicePollUC: devicePollUC,
	}
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

// RegisterRequest is the body for POST /auth/register.
type RegisterRequest struct {
	Body struct {
		Name     string `json:"name" required:"true" maxLength:"255" doc:"Display name"`
		Email    string `json:"email" required:"true" maxLength:"255" doc:"Email address"`
		Username string `json:"username,omitempty" maxLength:"100" doc:"Desired username (auto-generated from name if omitted)"`
	}
}

// RegisterResponse is the response for POST /auth/register.
type RegisterResponse struct {
	Body struct {
		Message string `json:"message" doc:"Registration confirmation with next steps"`
		UserID  string `json:"user_id" doc:"The Finsplitter user ID"`
	}
}

// Register handles user registration.
// POST /auth/register.
func (h *Handler) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	output, err := h.registerUC.Execute(ctx, auth.RegisterInput{
		Name:     req.Body.Name,
		Email:    req.Body.Email,
		Username: req.Body.Username,
	})
	if err != nil {
		if errors.Is(err, errs.ErrUsernameTaken) {
			return nil, huma.Error409Conflict("username already taken")
		}
		if errors.Is(err, errs.ErrUserAlreadyExists) {
			return nil, huma.Error409Conflict("user already registered")
		}
		return nil, huma.Error500InternalServerError("registration failed")
	}

	resp := &RegisterResponse{}
	resp.Body.Message = "Account created. Use POST /auth/device/auth to receive a verification code."
	resp.Body.UserID = output.UserID
	return resp, nil
}

// ---------------------------------------------------------------------------
// Me
// ---------------------------------------------------------------------------

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
