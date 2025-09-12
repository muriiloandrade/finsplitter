package cardbrand

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	slogctx "github.com/veqryn/slog-context"
)

type UpdateCardBrandRequest struct {
	ID   uuid.UUID `path:"id"`
	Body struct {
		Name string `example:"Visa" pattern:"^[a-zA-Z ]{1,50}$"`
	} `          doc:"Card brand details to update"`
}

type UpdateCardBrandResponse struct {
	Body entity.CardBrand `json:"data"`
}

type UpdateCardBrandHandler struct {
	UseCase usecases.UpdateCardBrandUC
}

func NewUpdateCardBrandHandler(uc usecases.UpdateCardBrandUC) UpdateCardBrandHandler {
	return UpdateCardBrandHandler{UseCase: uc}
}

func (h UpdateCardBrandHandler) UpdateCardBrand(
	ctx context.Context,
	input *UpdateCardBrandRequest,
) (*UpdateCardBrandResponse, error) {
	logger := slogctx.FromCtx(ctx)
	brand, err := h.UseCase.UpdateCardBrand(ctx, ports.UpdateCardBrandOptions{
		Id:   input.ID,
		Name: input.Body.Name,
	})
	if err != nil {
		logger.Error("Failed to update card brand", slog.Any("error", err))

		if errors.Is(err, errs.ErrCardBrandNotFound) {
			return nil, huma.Error404NotFound("CardBrand not found")
		}
		if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
			return nil, huma.Error409Conflict(err.Error())
		}

		return nil, huma.Error500InternalServerError(err.Error())
	}
	return &UpdateCardBrandResponse{Body: *brand}, nil
}
