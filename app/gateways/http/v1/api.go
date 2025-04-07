package v1

import (
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

type interfaceAPI interface {
	RegisterRoutes(r *chi.Mux, api huma.API, logger *slog.Logger)
}

type API struct {
	LivenessHandler  http.HandlerFunc
	ReadinessHandler http.HandlerFunc
	Logger           *slog.Logger
}

func (a API) Routes(r *chi.Mux) huma.API {
	var api huma.API

	r.Get("/health/liveness", a.LivenessHandler)
	r.Get("/health/readiness", a.ReadinessHandler)

	api = humachi.New(r, NewOpenAPIConfig())

	return api
}
