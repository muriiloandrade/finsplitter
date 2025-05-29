package usecases

import (
	"context"
	"errors"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

type CreateCardBrandUC struct {
	tx   domain.Transactioner
	repo ports.CreateCardBrandRepository
}

func NewCreateCardBrandUC(repo ports.CreateCardBrandRepository, tx domain.Transactioner) CreateCardBrandUC {
	return CreateCardBrandUC{repo: repo, tx: tx}
}

func (uc *CreateCardBrandUC) CreateCardBrand(ctx context.Context, name string) (*entity.CardBrand, error) {
	var insertedCardBrand *entity.CardBrand

	if name == "" {
		return nil, errors.New("name is required")
	}

	if err := uc.tx.WithTx(ctx, func(ctx context.Context) error {
		cb, err := uc.repo.CreateCardBrand(ctx, name)

		if err != nil {
			if errors.Is(err, errs.ErrCardBrandNotFound) {
				return errs.ErrCardBrandNotFound
			}
			if errors.Is(err, errs.ErrCardBrandAlreadyExists) {
				return errs.ErrCardBrandAlreadyExists
			}
		}

		insertedCardBrand = cb
		return err
	}); err != nil {
		return nil, err
	}

	return insertedCardBrand, nil
}
