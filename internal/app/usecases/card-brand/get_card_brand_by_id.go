package usecases

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

type GetCardBrandByIDUC struct {
	repo ports.GetCardBrandByIDRepository
}

func NewGetCardBrandByIDUC(repo ports.GetCardBrandByIDRepository) GetCardBrandByIDUC {
	return GetCardBrandByIDUC{repo: repo}
}

func (uc *GetCardBrandByIDUC) GetCardBrandByID(
	ctx context.Context,
	id uuid.UUID,
) (*entity.CardBrand, error) {
	brand, err := uc.repo.GetCardBrandByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrCardBrandNotFound) {
			return nil, errs.ErrCardBrandNotFound
		}
		return nil, err
	}
	return brand, nil
}
