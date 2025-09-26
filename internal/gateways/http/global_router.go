package http

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(_ *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.SupressNotFound(r),
		middleware.CleanPath,
		middleware.Recoverer,
	)

	return r
}
