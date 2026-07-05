package cardbrand_test

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/muriiloandrade/finsplitter/api"
	"github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/card-brand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegisterRoutes_ContainsExpectedPaths verifies that RegisterRoutes
// registers all five CRUD paths in the OpenAPI spec.
func TestRegisterRoutes_ContainsExpectedPaths(t *testing.T) {
	r := chi.NewRouter()
	cfg := api.NewOpenAPIConfig()
	humaAPI := humachi.New(r, cfg)

	// Wire up dummy handlers — we never call them, we only inspect the
	// resulting OpenAPI spec to verify paths are registered.
	a := cardbrand.API{
		GetCardBrandHandler: func(
			_ context.Context,
			_ *cardbrand.GetCardBrandRequest,
		) (*cardbrand.GetCardBrandResponse, error) {
			return &cardbrand.GetCardBrandResponse{}, nil
		},
		ListCardBrandsHandler: func(
			_ context.Context,
			_ *cardbrand.ListCardBrandsRequest,
		) (*cardbrand.ListCardBrandsResponse, error) {
			return &cardbrand.ListCardBrandsResponse{}, nil
		},
		CreateCardBrandHandler: func(
			_ context.Context,
			_ *cardbrand.CreateCardBrandRequest,
		) (*cardbrand.CreateCardBrandResponse, error) {
			return &cardbrand.CreateCardBrandResponse{}, nil
		},
		UpdateCardBrandHandler: func(
			_ context.Context,
			_ *cardbrand.UpdateCardBrandRequest,
		) (*cardbrand.UpdateCardBrandResponse, error) {
			return &cardbrand.UpdateCardBrandResponse{}, nil
		},
		DeleteCardBrandHandler: func(
			_ context.Context,
			_ *cardbrand.DeleteCardBrandRequest,
		) (*cardbrand.DeleteCardBrandResponse, error) {
			return &cardbrand.DeleteCardBrandResponse{}, nil
		},
	}
	a.RegisterRoutes(humaAPI)

	oapi := humaAPI.OpenAPI()

	// Verify /card-brands path (GET list + POST create).
	pathItem := oapi.Paths["/card-brands"]
	require.NotNil(t, pathItem, "expected /card-brands path in OpenAPI spec")
	assert.NotNil(t, pathItem.Get, "expected GET /card-brands (list)")
	assert.NotNil(t, pathItem.Post, "expected POST /card-brands (create)")

	// Verify /card-brands/{id} path (GET get + PATCH update + DELETE delete).
	idPathItem := oapi.Paths["/card-brands/{id}"]
	require.NotNil(t, idPathItem, "expected /card-brands/{id} path in OpenAPI spec")
	assert.NotNil(t, idPathItem.Get, "expected GET /card-brands/{id} (get by ID)")
	assert.NotNil(t, idPathItem.Patch, "expected PATCH /card-brands/{id} (update)")
	assert.NotNil(t, idPathItem.Delete, "expected DELETE /card-brands/{id} (delete)")
}
