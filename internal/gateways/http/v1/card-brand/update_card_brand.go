package cardbrand

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

const updateOperation = "handler.UpdateCardBrand"

type UpdateCardBrandRequest struct {
	Path struct {
		ID uuid.UUID `path:"id"`
	}
	Body struct {
		Name string `json:"name" validate:"required,min=1"`
	}
}

type UpdateCardBrandResponse struct {
	Body entity.CardBrand `json:"data"`
}

type UpdateCardBrandHandler struct {
	useCase usecases.UpdateCardBrandUC
}

func NewUpdateCardBrandHandler(useCase usecases.UpdateCardBrandUC) UpdateCardBrandHandler {
	return UpdateCardBrandHandler{useCase: useCase}
}

func (h UpdateCardBrandHandler) UpdateCardBrand(ctx context.Context, input *UpdateCardBrandRequest) (*UpdateCardBrandResponse, error) {
	logger := slogctx.FromCtx(ctx)
	brand, err := h.useCase.UpdateCardBrand(ctx, input.Path.ID, input.Body.Name)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, huma.Error404NotFound("Card brand not found")
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, huma.Error409Conflict("Card brand name already exists")
			}
		}
		logger.Error("Failed to update card brand", slog.String("operation", updateOperation), slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &UpdateCardBrandResponse{Body: brand}, nil
}
