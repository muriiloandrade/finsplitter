package user

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

type UpdateUserRequest struct {
	Path struct {
		ID uuid.UUID `path:"id"`
	}
	Body struct {
		Name         string  `json:"name" validate:"required,min=1"`
		Email        string  `json:"email" validate:"required,email"`
		PhoneNumber  *string `json:"phoneNumber,omitempty"`
		Username     string  `json:"username" validate:"required,min=1"`
		PasswordHash string  `json:"passwordHash" validate:"required,min=1"`
	}
}

type UpdateUserResponse struct {
	Body entity.User `json:"data"`
}

type UpdateUserHandler struct {
	useCase usecases.UpdateUserUC
}

func NewUpdateUserHandler(useCase usecases.UpdateUserUC) UpdateUserHandler {
	return UpdateUserHandler{useCase: useCase}
}

func (h UpdateUserHandler) UpdateUser(ctx context.Context, input *UpdateUserRequest) (*UpdateUserResponse, error) {
	logger := slogctx.FromCtx(ctx)
	user := entity.User{
		Name:         input.Body.Name,
		Email:        input.Body.Email,
		PhoneNumber:  input.Body.PhoneNumber,
		Username:     input.Body.Username,
		PasswordHash: input.Body.PasswordHash,
	}
	updated, err := h.useCase.UpdateUser(ctx, input.Path.ID, user)
	if err != nil {
		logger.Error("Failed to update user", slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &UpdateUserResponse{Body: updated}, nil
}
