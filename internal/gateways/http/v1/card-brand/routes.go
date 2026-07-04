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
		Method:  http.MethodGet,
		Path:    "/card-brands",
		Summary: "List all card brands",
		Description: "Returns an array of all registered card brands sorted by name ascending. " +
			"Each brand includes its unique ID, the display name, and a timestamp of when it was created.",
		Tags:     []string{"Card Brand"},
		Security: []map[string][]string{{"bearerAuth": {}}},
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.ListCardBrandsHandler)

	huma.Register(api, huma.Operation{
		Method:  http.MethodPost,
		Path:    "/card-brands",
		Summary: "Create a card brand",
		Description: "Creates a new card brand with the given name. " +
			"The name must be unique across all brands — a 409 Conflict is returned if it already exists. " +
			"Returns the created brand with its server-generated ID.",
		Tags:     []string{"Card Brand"},
		Security: []map[string][]string{{"bearerAuth": {}}},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusInternalServerError,
		},
	}, a.CreateCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/card-brands/{id}",
		Summary: "Get a card brand by ID",
		Description: "Retrieves a single card brand by its unique identifier (UUID). " +
			"Returns 404 Not Found if no brand matches the given ID.",
		Tags:     []string{"Card Brand"},
		Security: []map[string][]string{{"bearerAuth": {}}},
		Errors: []int{
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, a.GetCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:  http.MethodPatch,
		Path:    "/card-brands/{id}",
		Summary: "Update a card brand",
		Description: "Updates the name of an existing card brand identified by its UUID. " +
			"The new name must not conflict with another existing brand. " +
			"Returns the updated brand object on success.",
		Tags:     []string{"Card Brand"},
		Security: []map[string][]string{{"bearerAuth": {}}},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, a.UpdateCardBrandHandler)

	huma.Register(api, huma.Operation{
		Method:  http.MethodDelete,
		Path:    "/card-brands/{id}",
		Summary: "Delete a card brand",
		Description: "Permanently removes a card brand by its unique identifier (UUID). " +
			"Returns 404 if no brand matches the given ID. " +
			"Returns 409 Conflict if the brand is referenced by other entities (e.g. transactions, cards).",
		Tags:     []string{"Card Brand"},
		Security: []map[string][]string{{"bearerAuth": {}}},
		Errors: []int{
			http.StatusNotFound,
			http.StatusInternalServerError,
		},
	}, a.DeleteCardBrandHandler)
}
