package usecases

import (
	"context"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

type DeleteCardBrandUC struct {
	tx   domain.Transactioner
	repo ports.DeleteCardBrandRepository
}

func NewDeleteCardBrandUC(
	repo ports.DeleteCardBrandRepository,
	tx domain.Transactioner,
) DeleteCardBrandUC {
	return DeleteCardBrandUC{repo: repo, tx: tx}
}

func (uc *DeleteCardBrandUC) DeleteCardBrand(
	ctx context.Context,
	id uuid.UUID,
) (*entity.CardBrand, error) {
	var cardBrand *entity.CardBrand

	if id.IsNil() {
		return nil, errors.New("id is required")
	}

	err := uc.tx.WithTx(ctx, func(ctx context.Context) error {
		brand, err := uc.repo.DeleteCardBrand(ctx, id)
		if err != nil {
			return err
		}

		cardBrand = brand
		return nil
	})
	if err != nil {
		if errors.Is(err, errs.ErrCardBrandNotFound) {
			return nil, errs.ErrCardBrandNotFound
		}
		return nil, err
	}

	return cardBrand, nil
}
