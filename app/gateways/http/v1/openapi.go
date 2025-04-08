package v1

import "github.com/danielgtaylor/huma/v2"

func NewOpenAPIConfig() huma.Config {
	schemaPrefix := "#/components/schemas/"
	schemasPath := "/schemas"

	registry := huma.NewMapRegistry(schemaPrefix, huma.DefaultSchemaNamer)

	linkTransformer := huma.NewSchemaLinkTransformer(schemaPrefix, schemasPath)

	cfg := huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:       "Finsplitter",
				Version:     "0.0.1",
				Description: "Finsplitter is my personal finance control app",
				Contact: &huma.Contact{
					Name:  "Murilo A.",
					URL:   "muriloandrade.dev",
					Email: "murilo@muriloandrade.dev",
				},
			},
			Components:     &huma.Components{Schemas: registry},
			OnAddOperation: []huma.AddOpFunc{linkTransformer.OnAddOperation},
			Servers: []*huma.Server{
				{URL: "http://localhost:3033"},
			},
		},
		OpenAPIPath: "/openapi",
		DocsPath:    "/docs",
		SchemasPath: schemasPath,
		Formats: map[string]huma.Format{
			"application/json": huma.DefaultJSONFormat,
			"json":             huma.DefaultJSONFormat,
		},
		DefaultFormat: "application/json",
		Transformers:  []huma.Transformer{},
	}

	return cfg
}
