package profile

import (
	"context"
	"testing"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/muriiloandrade/finsplitter/api"
)

func TestRegisterRoutes_ContainsProfileSetupPath(t *testing.T) {
	r := chi.NewRouter()
	cfg := api.NewOpenAPIConfig()
	humaAPI := humachi.New(r, cfg)

	a := API{
		SetupHandler: func(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
			return nil, nil
		},
	}
	a.RegisterRoutes(humaAPI)

	oapi := humaAPI.OpenAPI()
	pathItem := oapi.Paths["/profile/setup"]
	require.NotNil(t, pathItem, "expected /profile/setup path in OpenAPI spec")
	require.NotNil(t, pathItem.Patch, "expected PATCH method on /profile/setup")
	assert.Equal(t, "Complete initial profile setup", pathItem.Patch.Summary)
	assert.ElementsMatch(t, []string{"Profile"}, pathItem.Patch.Tags)
}
