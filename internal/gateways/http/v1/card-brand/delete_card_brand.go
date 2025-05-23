package cardbrand

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases"
	slogctx "github.com/veqryn/slog-context"
)

const deleteOperation = "handler.DeleteCardBrand"

type DeleteCardBrandRequest struct {
	Path struct {
		ID uuid.UUID `path:"id"`
	}
}

type DeleteCardBrandResponse struct{}

type DeleteCardBrandHandler struct {
	useCase usecases.DeleteCardBrandUC
}

func NewDeleteCardBrandHandler(useCase usecases.DeleteCardBrandUC) DeleteCardBrandHandler {
	return DeleteCardBrandHandler{useCase: useCase}
}

func (h DeleteCardBrandHandler) DeleteCardBrand(ctx context.Context, input *DeleteCardBrandRequest) (*DeleteCardBrandResponse, error) {
	logger := slogctx.FromCtx(ctx)
	err := h.useCase.DeleteCardBrand(ctx, input.Path.ID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound("Card brand not found")
		}
		logger.Error("Failed to delete card brand", slog.String("operation", deleteOperation), slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &DeleteCardBrandResponse{}, nil
}
