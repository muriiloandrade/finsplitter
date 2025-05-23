package user

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

type ListUsersRequest struct{}

type ListUsersResponse struct {
	Body struct {
		Users      []entity.User `json:"data"`
		TotalCount int           `json:"total_count"`
	}
}

type ListUsersHandler struct {
	useCase usecases.ListUserUC
}

func NewListUsersHandler(useCase usecases.ListUserUC) ListUsersHandler {
	return ListUsersHandler{useCase: useCase}
}

func (h ListUsersHandler) ListUsers(ctx context.Context, input *ListUsersRequest) (*ListUsersResponse, error) {
	logger := slogctx.FromCtx(ctx)
	users, err := h.useCase.ListUsers(ctx)
	if err != nil {
		logger.Error("Failed to list users", slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &ListUsersResponse{
		Body: struct {
			Users      []entity.User `json:"data"`
			TotalCount int           `json:"total_count"`
		}{
			Users:      users,
			TotalCount: len(users),
		},
	}, nil
}
