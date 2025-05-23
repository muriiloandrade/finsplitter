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
	// GetCardBrandHandler    Handler[GetCardBrandRequest, GetCardBrandResponse]
	ListCardBrandsHandler  Handler[ListCardBrandsRequest, ListCardBrandsResponse]
	CreateCardBrandHandler Handler[CreateCardBrandRequest, CreateCardBrandResponse]
	UpdateCardBrandHandler Handler[UpdateCardBrandRequest, UpdateCardBrandResponse]
	DeleteCardBrandHandler Handler[DeleteCardBrandRequest, DeleteCardBrandResponse]
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

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/card-brands",
		Description: "Create card brand",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.CreateCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		Path:        "/card-brands/{id}",
		Description: "Update card brand",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusConflict,
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.UpdateCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodDelete,
		Path:        "/card-brands/{id}",
		Description: "Delete card brand",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.DeleteCardBrandHandler)
}
