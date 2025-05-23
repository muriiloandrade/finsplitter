package user

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

type CreateUserRequest struct {
	Body struct {
		Name         string  `json:"name" validate:"required,min=1"`
		Email        string  `json:"email" validate:"required,email"`
		PhoneNumber  *string `json:"phoneNumber,omitempty"`
		Username     string  `json:"username" validate:"required,min=1"`
		PasswordHash string  `json:"passwordHash" validate:"required,min=1"`
	}
}

type CreateUserResponse struct {
	Body entity.User `json:"data"`
}

type CreateUserHandler struct {
	useCase usecases.CreateUserUC
}

func NewCreateUserHandler(useCase usecases.CreateUserUC) CreateUserHandler {
	return CreateUserHandler{useCase: useCase}
}

func (h CreateUserHandler) CreateUser(ctx context.Context, input *CreateUserRequest) (*CreateUserResponse, error) {
	logger := slogctx.FromCtx(ctx)
	user := entity.User{
		Name:         input.Body.Name,
		Email:        input.Body.Email,
		PhoneNumber:  input.Body.PhoneNumber,
		Username:     input.Body.Username,
		PasswordHash: input.Body.PasswordHash,
	}
	created, err := h.useCase.CreateUser(ctx, user)
	if err != nil {
		logger.Error("Failed to create user", slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &CreateUserResponse{Body: created}, nil
}
