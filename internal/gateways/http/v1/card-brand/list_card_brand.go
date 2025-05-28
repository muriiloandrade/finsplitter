package cardbrand

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"

	slogctx "github.com/veqryn/slog-context"
)

const operation = "handler.ListCardBrands"

type ListCardBrandsRequest struct{}

type ListCardBrandsResponse struct {
	Body struct {
		CardBrands []entity.CardBrand `json:"data"`
		TotalCount int                `json:"total_count"`
	}
}

type ListCardBrandsHandler struct {
	useCase usecases.ListCardBrandUC
}

func NewListCardBrandsHandler(useCase usecases.ListCardBrandUC) ListCardBrandsHandler {
	return ListCardBrandsHandler{
		useCase: useCase,
	}
}

func (h ListCardBrandsHandler) ListCardBrands(ctx context.Context, input *ListCardBrandsRequest) (*ListCardBrandsResponse, error) {
	logger := slogctx.FromCtx(ctx)

	cardBrands, err := h.useCase.ListCardBrands(ctx)
	if err != nil {
		logger.Error("Failed to list card brands", slog.String("operation", operation), slog.Any("error", err))
		return nil, huma.Error500InternalServerError(err.Error())
	}

	return &ListCardBrandsResponse{
		Body: struct {
			CardBrands []entity.CardBrand `json:"data"`
			TotalCount int                `json:"total_count"`
		}{
			CardBrands: cardBrands,
			TotalCount: len(cardBrands),
		},
	}, nil
}
