package cardbrand

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	slogctx "github.com/veqryn/slog-context"
)

type GetCardBrandRequest struct {
	ID uuid.UUID `path:"id" doc:"Card brand ID" format:"uuid"`
}

type GetCardBrandResponse struct {
	Body entity.CardBrand `json:"data"`
}

type GetCardBrandHandler struct {
	UseCase usecases.GetCardBrandByIDUseCase
}

func NewGetCardBrandHandler(uc usecases.GetCardBrandByIDUseCase) GetCardBrandHandler {
	return GetCardBrandHandler{UseCase: uc}
}

func (h GetCardBrandHandler) GetCardBrand(
	ctx context.Context,
	input *GetCardBrandRequest,
) (*GetCardBrandResponse, error) {
	logger := slogctx.FromCtx(ctx)
	brand, err := h.UseCase.GetCardBrandByID(ctx, input.ID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get card brand", slog.Any("error", err))
		if errors.Is(err, errs.ErrCardBrandNotFound) {
			return nil, huma.Error404NotFound("CardBrand not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &GetCardBrandResponse{Body: *brand}, nil
}
