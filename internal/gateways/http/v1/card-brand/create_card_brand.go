package cardbrand

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/pkg/logctx"
)

type CreateCardBrandRequest struct {
	Body struct {
		Name string `json:"name" required:"true" example:"Visa"`
	}
}

type CreateCardBrandResponse struct {
	Body entity.CardBrand `json:"data"`
}

type CreateCardBrandHandler struct {
	UseCase usecases.CreateCardBrandUseCase
}

func NewCreateCardBrandHandler(uc usecases.CreateCardBrandUseCase) CreateCardBrandHandler {
	return CreateCardBrandHandler{UseCase: uc}
}

func (h CreateCardBrandHandler) CreateCardBrand(
	ctx context.Context,
	input *CreateCardBrandRequest,
) (*CreateCardBrandResponse, error) {
	logger := logctx.FromCtx(ctx)
	brand, err := h.UseCase.CreateCardBrand(ctx, input.Body.Name)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create card brand", slog.Any("error", err))
		if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
			return nil, huma.Error409Conflict(err.Error())
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &CreateCardBrandResponse{*brand}, nil
}
