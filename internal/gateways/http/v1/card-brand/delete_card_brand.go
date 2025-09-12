package cardbrand

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	slogctx "github.com/veqryn/slog-context"
)

type DeleteCardBrandRequest struct {
	ID uuid.UUID `path:"id"`
}

type DeleteCardBrandResponse struct {
	Body struct {
		Deleted bool `json:"deleted"`
	}
}

type DeleteCardBrandHandler struct {
	UseCase usecases.DeleteCardBrandUC
}

func NewDeleteCardBrandHandler(uc usecases.DeleteCardBrandUC) DeleteCardBrandHandler {
	return DeleteCardBrandHandler{UseCase: uc}
}

func (h DeleteCardBrandHandler) DeleteCardBrand(
	ctx context.Context,
	input *DeleteCardBrandRequest,
) (*DeleteCardBrandResponse, error) {
	logger := slogctx.FromCtx(ctx)
	cb, err := h.UseCase.DeleteCardBrand(ctx, input.ID)
	if err != nil {
		logger.Error("Failed to delete card brand", slog.Any("error", err))
		if errors.Is(err, errs.ErrCardBrandNotFound) {
			return nil, huma.Error404NotFound("CardBrand not found")
		}
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &DeleteCardBrandResponse{Body: struct {
		Deleted bool `json:"deleted"`
	}{Deleted: cb != nil}}, nil
}
