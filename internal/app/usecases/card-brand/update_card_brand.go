package usecases

import (
	"context"
	"errors"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

type UpdateCardBrandUC struct {
	tx   domain.Transactioner
	repo ports.UpdateCardBrandRepository
}

func NewUpdateCardBrandUC(
	repo ports.UpdateCardBrandRepository,
	tx domain.Transactioner,
) UpdateCardBrandUC {
	return UpdateCardBrandUC{repo: repo, tx: tx}
}

func (uc *UpdateCardBrandUC) UpdateCardBrand(
	ctx context.Context,
	opts ports.UpdateCardBrandOptions,
) (*entity.CardBrand, error) {
	var cardBrand *entity.CardBrand

	if opts.Name == "" || opts.Id.IsNil() {
		return nil, errors.New("name and id are required")
	}

	err := uc.tx.WithTx(ctx, func(ctx context.Context) error {
		brand, err := uc.repo.UpdateCardBrand(ctx, opts)
		if err != nil {
			if errors.Is(err, errs.ErrCardBrandNotFound) {
				return errs.ErrCardBrandNotFound
			}
			if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
				return errs.ErrCardBrandAlreadyExists
			}
			return err
		}

		cardBrand = brand
		return nil
	})
	if err != nil {
		return nil, err
	}

	return cardBrand, nil
}
