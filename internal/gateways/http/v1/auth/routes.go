package auth

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases/auth"
	"github.com/muriiloandrade/finsplitter/internal/gateways/logto"
)

// HumaHandler defines the function signature for Huma handlers.
type HumaHandler[I, O any] func(context.Context, *I) (*O, error)

// API holds the auth handler references and registers routes.
type API struct {
	RegisterHandler HumaHandler[RegisterRequest, RegisterResponse]
	MeHandler       HumaHandler[struct{}, MeResponse]
	SignInHandler   HumaHandler[SignInRequest, SignInResponse]
}

// NewAPI creates an auth API from the given dependencies.
func NewAPI(userRepo ports.UserRepository, logtoM2M *logto.Client) API {
	registerUC := auth.NewRegisterUseCase(userRepo, logtoM2M)
	meUC := auth.NewMeUseCase(userRepo)
	signInUC := auth.NewSignInUseCase(logtoM2M)
	h := NewHandler(registerUC, meUC, signInUC)

	return API{
		RegisterHandler: h.Register,
		MeHandler:       h.Me,
		SignInHandler:   h.SignIn,
	}
}

// RegisterRoutes registers auth routes on the given Huma API.
func (a API) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Description: "Register a new user in Logto and Finsplitter",
		Tags:        []string{"Auth"},
		Errors: []int{
			http.StatusConflict,
			http.StatusInternalServerError,
		},
	}, a.RegisterHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Description: "Get current user profile from JWT claims",
		Tags:        []string{"Auth"},
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusForbidden,
		},
	}, a.MeHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/auth/sign-in",
		Description: "Sign in with email and password, returns OIDC tokens",
		Tags:        []string{"Auth"},
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.SignInHandler)
}
