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
		expectedError error
		wantCardBrand *entity.CardBrand
	}{
		{
			name:  "success - creates card brand",
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
			expectedError: nil,
			wantCardBrand: validCardBrand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockCreateCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)

			uc := NewCreateCardBrandUC(repo, tx)

			got, err := uc.CreateCardBrand(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCardBrand, got)
			}

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
		wantCardBrand *entity.CardBrand
	}{
		{
			name:          "empty name",
			input:         "",
			repoSetup:     func(repo *ports.MockCreateCardBrandRepository) {},
			txSetup:       func(tx *domain.MockTransactioner) {},
			expectedError: errors.New("name is required"),
			wantCardBrand: nil,
		},
		{
			name:  "card brand already exists",
			input: "Visa",
			repoSetup: func(repo *ports.MockCreateCardBrandRepository) {
				repo.EXPECT().CreateCardBrand(mock.Anything, "Visa").Return(nil, errs.ErrCardBrandAlreadyExists)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).Return(errs.ErrCardBrandAlreadyExists)
			},
			expectedError: errs.ErrCardBrandAlreadyExists,
			wantCardBrand: nil,
		},
		{
			name:  "database generic error",
			input: "Visa",
			repoSetup: func(repo *ports.MockCreateCardBrandRepository) {
				repo.EXPECT().CreateCardBrand(mock.Anything, "Visa").Return(nil, errs.ErrDatabaseGeneric)
			},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Run(func(ctx context.Context, fn domain.TransactionFunc) {
						fn(ctx)
					}).Return(errs.ErrDatabaseGeneric)
			},
			expectedError: errs.ErrDatabaseGeneric,
			wantCardBrand: nil,
		},
		{
			name:      "transaction failed",
			input:     "Visa",
			repoSetup: func(repo *ports.MockCreateCardBrandRepository) {},
			txSetup: func(tx *domain.MockTransactioner) {
				tx.EXPECT().WithTx(mock.Anything, mock.AnythingOfType("domain.TransactionFunc")).
					Return(errors.New("transaction failed"))
			},
			expectedError: errors.New("transaction failed"),
			wantCardBrand: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockCreateCardBrandRepository(t)
			tx := domain.NewMockTransactioner(t)
			tt.repoSetup(repo)
			tt.txSetup(tx)

			uc := NewCreateCardBrandUC(repo, tx)

			got, err := uc.CreateCardBrand(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCardBrand, got)
			}

			repo.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
