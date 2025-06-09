package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	"github.com/muriiloandrade/finsplitter/internal/domain"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
)

func TestDeleteCardBrandUC_DeleteCardBrandSuccess(t *testing.T) {
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
		inputID   uuid.UUID
		repoSetup func(repo *ports.MockDeleteCardBrandRepository)
		txSetup   func(tx *domain.MockTransactioner)
		want      *entity.CardBrand
	}{
		{
			name:    "deletes card brand",
			inputID: id,
			repoSetup: func(repo *ports.MockDeleteCardBrandRepository) {
				repo.EXPECT().DeleteCardBrand(mock.Anything, id).Return(cardBrand, nil)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).Run(func(ctx context.Context, fn domain.TransactionFunc) {
					fn(ctx)
				}).Return(nil)
			},
			want: cardBrand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockDeleteCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)
			uc := NewDeleteCardBrandUC(repo, tx)
			got, err := uc.DeleteCardBrand(context.Background(), tt.inputID)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestDeleteCardBrandUC_DeleteCardBrandError(t *testing.T) {
	id := uuid.Must(uuid.NewV4())

	tests := []struct {
		name      string
		inputID   uuid.UUID
		repoSetup func(repo *ports.MockDeleteCardBrandRepository)
		txSetup   func(tx *domain.MockTransactioner)
		err       error
	}{
		{
			name:      "returns error on empty id",
			inputID:   uuid.Nil,
			repoSetup: func(repo *ports.MockDeleteCardBrandRepository) {},
			txSetup:   func(tx *domain.MockTransactioner) {},
			err:       errors.New("id is required"),
		},
		{
			name:    "returns error on id not found",
			inputID: id,
			repoSetup: func(repo *ports.MockDeleteCardBrandRepository) {
				repo.EXPECT().DeleteCardBrand(mock.Anything, id).Return(nil, errs.ErrCardBrandNotFound)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).Run(func(ctx context.Context, fn domain.TransactionFunc) {
					fn(ctx)
				}).Return(errs.ErrCardBrandNotFound)
			},
			err: errs.ErrCardBrandNotFound,
		},
		{
			name:    "returns error on foreign key violation",
			inputID: id,
			repoSetup: func(repo *ports.MockDeleteCardBrandRepository) {
				repo.EXPECT().DeleteCardBrand(mock.Anything, id).Return(nil, errs.ErrCardBrandForeignKeyViolation)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).Run(func(ctx context.Context, fn domain.TransactionFunc) {
					fn(ctx)
				}).Return(errs.ErrCardBrandForeignKeyViolation)
			},
			err: errs.ErrCardBrandForeignKeyViolation,
		},
		{
			name:      "returns error on transaction failed",
			inputID:   id,
			repoSetup: func(repo *ports.MockDeleteCardBrandRepository) {},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).Return(errors.New("transaction failed"))
			},
			err: errors.New("transaction failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockDeleteCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)
			uc := NewDeleteCardBrandUC(repo, tx)
			got, err := uc.DeleteCardBrand(context.Background(), tt.inputID)
			assert.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, tt.err, err)
			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
