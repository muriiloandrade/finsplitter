package v1

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	openapi "github.com/muriiloandrade/finsplitter/api"
	"github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/auth"
)

type interfaceAPI interface {
	RegisterRoutes(api huma.API)
}

// API aggregates all route groups and middleware for the v1 HTTP layer.
type API struct {
	LivenessHandler  http.HandlerFunc
	ReadinessHandler http.HandlerFunc
	CardBrandAPI     interfaceAPI
	AuthAPI          interfaceAPI
	ProfileAPI       interfaceAPI
	AuthMiddleware   *auth.Middleware
	Logger           *slog.Logger
}

// Routes builds the chi router and registers all route handlers via a single
// Huma v2 API instance. Protected routes use chi middleware with a skip list
// for public paths (docs, health, register), keeping OpenAPI docs public.
func (a API) Routes(r *chi.Mux) huma.API {
	// Apply auth middleware before any route registration (chi requirement).
	// The middleware internally skips /health/, /docs, /openapi, /auth/register,
	// and optionally populates claims for /auth/me.
	if a.AuthMiddleware != nil {
		r.Use(a.AuthMiddleware.Protected())
	}

	// Health endpoints — plain chi, no huma.
	r.Get("/health/liveness", a.LivenessHandler)
	r.Get("/health/readiness", a.ReadinessHandler)

	// Single huma API — docs are registered here and are public because
	// the middleware skips /docs, /openapi, etc.
	api := humachi.New(r, openapi.NewOpenAPIConfig())

	// Public & protected routes — protection is handled by chi middleware.
	a.CardBrandAPI.RegisterRoutes(api)
	a.AuthAPI.RegisterRoutes(api)
	a.ProfileAPI.RegisterRoutes(api)

	return api
}
