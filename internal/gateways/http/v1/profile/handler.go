package profile

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/profile"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"
)

// Handler handles profile-related HTTP requests.
type Handler struct {
	setupUC *profile.SetupUseCase
}

// NewHandler creates a new profile Handler.
func NewHandler(setupUC *profile.SetupUseCase) *Handler {
	return &Handler{
		setupUC: setupUC,
	}
}

// SetupRequest is the body for POST /profile/setup.
type SetupRequest struct {
	Body struct {
		Username string `json:"username" required:"true" maxLength:"100" doc:"Desired username"`
		Name     string `json:"name,omitempty" maxLength:"200" doc:"Display name"`
		Phone    string `json:"phone,omitempty" maxLength:"20" doc:"Phone number"`
		Picture  string `json:"picture,omitempty" maxLength:"2048" doc:"Avatar URL"`
	}
}

// SetupResponse is the response for POST /profile/setup.
type SetupResponse struct {
	Body struct {
		Message string `json:"message" doc:"Setup confirmation message"`
		UserID  string `json:"user_id" doc:"The Finsplitter user ID"`
	}
}

// Setup handles profile setup for first-time users.
// POST /profile/setup.
func (h *Handler) Setup(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("unauthenticated")
	}

	output, err := h.setupUC.Execute(ctx, profile.SetupInput{
		LogtoUserID: claims.Sub,
		Username:    req.Body.Username,
		Name:        req.Body.Name,
		Phone:       req.Body.Phone,
		Picture:     req.Body.Picture,
	})
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			return nil, huma.Error409Conflict("profile already set up")
		}
		if errors.Is(err, errs.ErrInvalidInput) {
			return nil, huma.Error422UnprocessableEntity("invalid input")
		}
		return nil, huma.Error500InternalServerError("setup failed")
	}

	resp := &SetupResponse{}
	resp.Body.UserID = output.UserID
	resp.Body.Message = "Profile setup complete. Please re-authenticate to receive an updated JWT."
	return resp, nil
}
