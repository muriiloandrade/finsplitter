package http

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
	otelchimetric "github.com/riandyrn/otelchi/metric"
	"go.opentelemetry.io/otel"
)

func NewRouter(_ *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	baseCfg := otelchimetric.NewBaseConfig(
		"finsplitter",
		otelchimetric.WithMeterProvider(otel.GetMeterProvider()),
	)

	r.Use(
		middleware.SupressNotFound(r),
		middleware.CleanPath,
		middleware.Recoverer,
		otelchi.Middleware(
			"finsplitter",
			otelchi.WithChiRoutes(r),
			otelchi.WithRequestMethodInSpanName(true),
			otelchi.WithTraceResponseHeaders(otelchi.TraceHeaderConfig{}),
			otelchi.WithFilter(func(r *http.Request) bool {
				blockList := []string{
					"/health/liveness",
					"/health/readiness",
					"/docs",
					"openapi.yml",
					"openapi.json",
				}
				return !slices.Contains(blockList, r.URL.Path)
			}),
		),
		otelchimetric.NewRequestDurationMillis(baseCfg),
		otelchimetric.NewRequestInFlight(baseCfg),
		otelchimetric.NewResponseSizeBytes(baseCfg),
	)

	return r
}
