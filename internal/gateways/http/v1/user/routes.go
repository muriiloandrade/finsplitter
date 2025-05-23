package user

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
)

type Handler[I, O any] func(context.Context, *I) (*O, error)

type API struct {
	ListUsersHandler  Handler[ListUsersRequest, ListUsersResponse]
	CreateUserHandler Handler[CreateUserRequest, CreateUserResponse]
	UpdateUserHandler Handler[UpdateUserRequest, UpdateUserResponse]
	DeleteUserHandler Handler[DeleteUserRequest, DeleteUserResponse]
}

func (a API) RegisterRoutes(r *chi.Mux, api huma.API, logger *slog.Logger) {
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/users",
		Description: "List users",
		Tags:        []string{"User"},
		Errors: []int{
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.ListUsersHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/users",
		Description: "Create user",
		Tags:        []string{"User"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusConflict,
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.CreateUserHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodPut,
		Path:        "/users/{id}",
		Description: "Update user",
		Tags:        []string{"User"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusConflict,
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.UpdateUserHandler)

	huma.Register(api, huma.Operation{
		Method:      http.MethodDelete,
		Path:        "/users/{id}",
		Description: "Delete user",
		Tags:        []string{"User"},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusUnauthorized,
			http.StatusInternalServerError,
		},
	}, a.DeleteUserHandler)
}
