package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateCardBrandUC_UpdateCardBrandSuccess(t *testing.T) {
	now := time.Now()
	id := uuid.Must(uuid.NewV4())
	cardBrand := &entity.CardBrand{
		ID:               id,
		Name:             "Visa",
		CreatedDate:      now,
		LastModifiedDate: now,
	}

	tests := []struct {
		name      string
		input     ports.UpdateCardBrandOptions
		repoSetup func(repo *ports.MockUpdateCardBrandRepository)
		txSetup   func(tx *domain.MockTransactioner)
		want      *entity.CardBrand
	}{
		{
			name:  "updates a card brand",
			input: ports.UpdateCardBrandOptions{ID: id, Name: "Visa"},
			repoSetup: func(repo *ports.MockUpdateCardBrandRepository) {
				repo.EXPECT().
					UpdateCardBrand(mock.Anything, ports.UpdateCardBrandOptions{ID: id, Name: "Visa"}).
					Return(cardBrand, nil)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().
					WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).
					Return(nil)
			},
			want: cardBrand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockUpdateCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)
			uc := usecases.NewUpdateCardBrandUC(repo, tx)
			got, err := uc.UpdateCardBrand(context.Background(), tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestUpdateCardBrandUC_UpdateCardBrandError(t *testing.T) {
	id := uuid.Must(uuid.NewV4())

	tests := []struct {
		name      string
		input     ports.UpdateCardBrandOptions
		repoSetup func(repo *ports.MockUpdateCardBrandRepository)
		txSetup   func(tx *domain.MockTransactioner)
		err       error
	}{
		{
			name:      "returns error on empty name",
			input:     ports.UpdateCardBrandOptions{ID: id, Name: ""},
			repoSetup: func(_ *ports.MockUpdateCardBrandRepository) {},
			txSetup:   func(_ *domain.MockTransactioner) {},
			err:       errors.New("name and id are required"),
		},
		{
			name:      "returns error on empty id",
			input:     ports.UpdateCardBrandOptions{ID: uuid.Nil, Name: "Visa"},
			repoSetup: func(_ *ports.MockUpdateCardBrandRepository) {},
			txSetup:   func(_ *domain.MockTransactioner) {},
			err:       errors.New("name and id are required"),
		},
		{
			name:  "returns error on not found",
			input: ports.UpdateCardBrandOptions{ID: id, Name: "Visa"},
			repoSetup: func(repo *ports.MockUpdateCardBrandRepository) {
				repo.EXPECT().
					UpdateCardBrand(mock.Anything, ports.UpdateCardBrandOptions{ID: id, Name: "Visa"}).
					Return(nil, errs.ErrCardBrandNotFound)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().
					WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).
					Return(errs.ErrCardBrandNotFound)
			},
			err: errs.ErrCardBrandNotFound,
		},
		{
			name:  "returns error on already exists",
			input: ports.UpdateCardBrandOptions{ID: id, Name: "Visa"},
			repoSetup: func(repo *ports.MockUpdateCardBrandRepository) {
				repo.EXPECT().
					UpdateCardBrand(mock.Anything, ports.UpdateCardBrandOptions{ID: id, Name: "Visa"}).
					Return(nil, errs.ErrCardBrandAlreadyExists)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().
					WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).
					Return(errs.ErrCardBrandAlreadyExists)
			},
			err: errs.ErrCardBrandAlreadyExists,
		},
		{
			name:      "returns error on transaction failed",
			input:     ports.UpdateCardBrandOptions{ID: id, Name: "Visa"},
			repoSetup: func(_ *ports.MockUpdateCardBrandRepository) {},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().
					WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Return(errors.New("transaction failed"))
			},
			err: errors.New("transaction failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockUpdateCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)
			uc := usecases.NewUpdateCardBrandUC(repo, tx)
			got, err := uc.UpdateCardBrand(context.Background(), tt.input)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, tt.err, err)
			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
