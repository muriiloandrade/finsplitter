package cardbrand

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type Handler[I, O any] func(context.Context, *I) (*O, error)

type API struct {
	GetCardBrandHandler    Handler[GetCardBrandRequest, GetCardBrandResponse]
	ListCardBrandsHandler  Handler[ListCardBrandsRequest, ListCardBrandsResponse]
	CreateCardBrandHandler Handler[CreateCardBrandRequest, CreateCardBrandResponse]
	UpdateCardBrandHandler Handler[UpdateCardBrandRequest, UpdateCardBrandResponse]
	DeleteCardBrandHandler Handler[DeleteCardBrandRequest, DeleteCardBrandResponse]
}

func (a API) RegisterRoutes(api huma.API) {
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
		Description: "Create a new card brand",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusInternalServerError,
		},
	}, a.CreateCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/card-brands/{id}",
		Description: "Get card brand by id",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, a.GetCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPatch,
		Path:        "/card-brands/{id}",
		Description: "Update card brand",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, a.UpdateCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodDelete,
		Path:        "/card-brands/{id}",
		Description: "Delete card brand",
		Tags:        []string{"Card Brand"},
		Errors: []int{
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, a.DeleteCardBrandHandler)
}
