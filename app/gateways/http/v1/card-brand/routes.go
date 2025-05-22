package cardbrand

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
)

type Handler[I, O any] func(context.Context, *I) (*O, error)

type API struct {
	// GetCardBrandHandler    Handler/[GetCardBrandRequest, GetCardBrandResponse]
	ListCardBrandsHandler Handler[ListCardBrandsRequest, ListCardBrandsResponse]
	// CreateCardBrandHandler Handler[CreateCardBrandRequest, CreateCardBrandResponse]
	// UpdateCardBrandHandler Handler[UpdateCardBrandRequest, UpdateCardBrandResponse]
	// DeleteCardBrandHandler Handler[DeleteCardBrandRequest, DeleteCardBrandResponse]
}

func (a API) RegisterRoutes(r *chi.Mux, api huma.API, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/card-brands",
		Description: "List card brands",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.ListCardBrandsHandler)
}
