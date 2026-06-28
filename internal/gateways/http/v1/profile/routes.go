package profile

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/profile"
)

// HumaHandler defines the function signature for Huma handlers.
type HumaHandler[I, O any] func(context.Context, *I) (*O, error)

// API holds the profile handler references and registers routes.
// Auth is enforced by the chi-level middleware; no per-operation middleware needed.
type API struct {
	SetupHandler HumaHandler[SetupRequest, SetupResponse]
}

// NewAPI creates a profile API from the given dependencies.
func NewAPI(userRepo ports.UserRepository) API {
	setupUC := profile.NewSetupUseCase(userRepo)
	h := NewHandler(setupUC)

	return API{
		SetupHandler: h.Setup,
	}
}

// RegisterRoutes registers profile routes on the given Huma API.
func (a API) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/profile/setup",
		Description: "Complete profile setup for a first-time user",
		Tags:        []string{"Profile"},
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusConflict,
			http.StatusUnprocessableEntity,
			http.StatusInternalServerError,
		},
	}, a.SetupHandler)
}
