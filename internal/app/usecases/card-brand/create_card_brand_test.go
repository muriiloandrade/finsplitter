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

func TestCreateCardBrandUC_CreateCardBrandSuccess(t *testing.T) {
	// Setup test data
	now := time.Now()
	sampleUUID := uuid.Must(uuid.NewV4())
	validCardBrand := &entity.CardBrand{
		ID:               sampleUUID,
		Name:             "Visa",
		CreatedDate:      now,
		LastModifiedDate: now,
	}

	tests := []struct {
		name          string
		input         string
		repoSetup     func(repo *ports.MockCreateCardBrandRepository)
		txSetup       func(tx *domain.MockTransactioner)
		wantCardBrand *entity.CardBrand
	}{
		{
			name:  "creates card brand",
			input: "Visa",
			repoSetup: func(repo *ports.MockCreateCardBrandRepository) {
				repo.EXPECT().CreateCardBrand(mock.Anything, "Visa").Return(validCardBrand, nil)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).Return(nil)
			},
			wantCardBrand: validCardBrand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockCreateCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)

			uc := usecases.NewCreateCardBrandUC(repo, tx)

			got, err := uc.CreateCardBrand(context.Background(), tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCardBrand, got)

			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}

func TestCreateCardBrandUC_CreateCardBrandError(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		repoSetup     func(repo *ports.MockCreateCardBrandRepository)
		txSetup       func(tx *domain.MockTransactioner)
		expectedError error
	}{
		{
			name:          "returns error on empty name",
			input:         "",
			repoSetup:     func(repo *ports.MockCreateCardBrandRepository) {},
			txSetup:       func(tx *domain.MockTransactioner) {},
			expectedError: errors.New("name is required"),
		},
		{
			name:  "returns error on card brand already exists",
			input: "Visa",
			repoSetup: func(repo *ports.MockCreateCardBrandRepository) {
				repo.EXPECT().
					CreateCardBrand(mock.Anything, "Visa").
					Return(nil, errs.ErrCardBrandAlreadyExists)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).Return(errs.ErrCardBrandAlreadyExists)
			},
			expectedError: errs.ErrCardBrandAlreadyExists,
		},
		{
			name:  "returns error on database generic error",
			input: "Visa",
			repoSetup: func(repo *ports.MockCreateCardBrandRepository) {
				repo.EXPECT().
					CreateCardBrand(mock.Anything, "Visa").
					Return(nil, errs.ErrDatabaseGeneric)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).Return(errs.ErrDatabaseGeneric)
			},
			expectedError: errs.ErrDatabaseGeneric,
		},
		{
			name:      "returns error on transaction failed",
			input:     "Visa",
			repoSetup: func(repo *ports.MockCreateCardBrandRepository) {},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Return(errors.New("transaction failed"))
			},
			expectedError: errors.New("transaction failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockCreateCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)

			uc := usecases.NewCreateCardBrandUC(repo, tx)

			got, err := uc.CreateCardBrand(context.Background(), tt.input)

			require.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
			assert.Nil(t, got)

			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
