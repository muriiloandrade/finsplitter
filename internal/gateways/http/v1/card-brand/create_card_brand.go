package cardbrand

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/muriiloandrade/finsplitter/internal/app/usecases"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	slogctx "github.com/veqryn/slog-context"
)

const createOperation = "handler.CreateCardBrand"

type CreateCardBrandRequest struct {
	Body struct {
		Name string `json:"name" validate:"required,min=1"`
	}
}

type CreateCardBrandResponse struct {
	Body entity.CardBrand `json:"data"`
}

type CreateCardBrandHandler struct {
	useCase usecases.CreateCardBrandUC
}

func NewCreateCardBrandHandler(useCase usecases.CreateCardBrandUC) CreateCardBrandHandler {
	return CreateCardBrandHandler{useCase: useCase}
}

func (h CreateCardBrandHandler) CreateCardBrand(ctx context.Context, input *CreateCardBrandRequest) (*CreateCardBrandResponse, error) {
	logger := slogctx.FromCtx(ctx)
	brand, err := h.useCase.CreateCardBrand(ctx, input.Body.Name)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, huma.Error409Conflict("Card brand name already exists")
			}
		}
		logger.Error("Failed to create card brand", slog.String("operation", createOperation), slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &CreateCardBrandResponse{Body: brand}, nil
}
