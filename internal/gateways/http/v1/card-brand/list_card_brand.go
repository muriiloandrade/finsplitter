package cardbrand

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"

	slogctx "github.com/veqryn/slog-context"
)

const operation = "handler.ListCardBrands"

type ListCardBrandsRequest struct {
	Id         uuid.UUID `query:"id" doc:"Card brand ID" nullable:"true" format:"uuid"`
	Name       string    `query:"name" doc:"Card brand name" nullable:"true" example:"Visa" pattern:"^[a-zA-Z ]{1,50}$"`
	PageSize   int       `query:"page_size" doc:"Number of items per page" default:"10" minimum:"1" maximum:"100"`
	PageNumber int       `query:"page_number" doc:"Page number for pagination" default:"1" minimum:"1"`
}

type ListCardBrandsResponse struct {
	Body struct {
		CardBrands []entity.CardBrand `json:"data"`
		TotalCount int                `json:"total_count"`
	}
}

type ListCardBrandsHandler struct {
	useCase usecases.ListCardBrandsUC
}

func NewListCardBrandsHandler(useCase usecases.ListCardBrandsUC) ListCardBrandsHandler {
	return ListCardBrandsHandler{
		useCase: useCase,
	}
}

func (h ListCardBrandsHandler) ListCardBrands(ctx context.Context, input *ListCardBrandsRequest) (*ListCardBrandsResponse, error) {
	logger := slogctx.FromCtx(ctx)

	cardBrands, err := h.useCase.ListCardBrands(ctx, ports.ListCardBrandFilterOptions{
		Id:         input.Id,
		Name:       &input.Name,
		PageSize:   input.PageSize,
		PageNumber: input.PageNumber,
	})
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
