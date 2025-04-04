package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
)

// CLIOptions for the CLI.
type CLIOptions struct {
	Port int `help:"Port to listen on" short:"p" default:"3033"`
}

// GreetingOutput represents the greeting operation response.
type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

type GreetingInput struct {
	Name string `path:"name" maxLength:"30" example:"world" doc:"Name to greet"`
}

func main() {
	// Create a CLI app
	cli := humacli.New(func(hooks humacli.Hooks, options *CLIOptions) {
		router := http.NewServeMux()

		cfg := huma.DefaultConfig("Finsplitter", "1.0.0")

		cfg.Servers = append(cfg.Servers, &huma.Server{
			URL:         "http://{host}",
			Description: "Local Server",
			Variables: map[string]*huma.ServerVariable{
				"host": {
					Default:     "localhost:3033",
					Description: "The host of the environment",
					Enum:        []string{"localhost:3033"},
				},
			},
		})
		cfg.Info.Description = "Finsplitter is my personal finance control app"
		cfg.Info.Contact = &huma.Contact{
			Name:  "Murilo A.",
			URL:   "muriloandrade.dev",
			Email: "murilo@muriloandrade.dev",
		}

		api := humago.New(router, cfg)

		// Register GET /greeting/{name}
		huma.Register(api, huma.Operation{
			OperationID: "get-greeting",
			Summary:     "Get a greeting",
			Method:      http.MethodGet,
			Path:        "/greeting/{name}",
		}, func(ctx context.Context, input *GreetingInput) (*GreetingOutput, error) {
			resp := &GreetingOutput{}
			resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)
			return resp, nil
		})

		// Create the HTTP server.
		server := http.Server{
			Addr:    fmt.Sprintf(":%d", options.Port),
			Handler: router,
		}

		// Tell the CLI how to start your router.
		hooks.OnStart(func() {
			log.Println("Starting server...")
			server.ListenAndServe()
		})

		// Tell the CLI how to stop your server.
		hooks.OnStop(func() {
			// Give the server 5 seconds to gracefully shut down, then give up.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
		})
	})

	// Run the CLI. When passed no commands, it starts the server.
	cli.Run()
}
