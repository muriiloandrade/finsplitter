package user

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases"
	slogctx "github.com/veqryn/slog-context"
)

type DeleteUserRequest struct {
	Path struct {
		ID uuid.UUID `path:"id"`
	}
}

type DeleteUserResponse struct{}

type DeleteUserHandler struct {
	useCase usecases.DeleteUserUC
}

func NewDeleteUserHandler(useCase usecases.DeleteUserUC) DeleteUserHandler {
	return DeleteUserHandler{useCase: useCase}
}

func (h DeleteUserHandler) DeleteUser(ctx context.Context, input *DeleteUserRequest) (*DeleteUserResponse, error) {
	logger := slogctx.FromCtx(ctx)
	err := h.useCase.DeleteUser(ctx, input.Path.ID)
	if err != nil {
		logger.Error("Failed to delete user", slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &DeleteUserResponse{}, nil
}
