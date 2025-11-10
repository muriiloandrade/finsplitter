package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetCardBrandByIDUC_GetCardBrandByIDSuccess(t *testing.T) {
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
		repoSetup func(repo *ports.MockGetCardBrandByIDRepository)
		want      *entity.CardBrand
	}{
		{
			name:    "returns card brand by id",
			inputID: id,
			repoSetup: func(repo *ports.MockGetCardBrandByIDRepository) {
				repo.EXPECT().GetCardBrandByID(mock.Anything, id).Return(cardBrand, nil)
			},
			want: cardBrand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockGetCardBrandByIDRepository(t)
			tt.repoSetup(repo)
			uc := usecases.NewGetCardBrandByIDUC(repo)
			got, err := uc.GetCardBrandByID(context.Background(), tt.inputID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetCardBrandByIDUC_GetCardBrandByIDError(t *testing.T) {
	id := uuid.Must(uuid.NewV4())

	tests := []struct {
		name      string
		inputID   uuid.UUID
		repoSetup func(repo *ports.MockGetCardBrandByIDRepository)
		err       error
	}{
		{
			name:    "returns error on id not found",
			inputID: id,
			repoSetup: func(repo *ports.MockGetCardBrandByIDRepository) {
				repo.EXPECT().
					GetCardBrandByID(mock.Anything, id).
					Return(nil, errs.ErrCardBrandNotFound)
			},
			err: errs.ErrCardBrandNotFound,
		},
		{
			name:    "returns error on database generic error",
			inputID: id,
			repoSetup: func(repo *ports.MockGetCardBrandByIDRepository) {
				repo.EXPECT().
					GetCardBrandByID(mock.Anything, id).
					Return(nil, errs.ErrDatabaseGeneric)
			},
			err: errs.ErrDatabaseGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockGetCardBrandByIDRepository(t)
			tt.repoSetup(repo)
			uc := usecases.NewGetCardBrandByIDUC(repo)
			got, err := uc.GetCardBrandByID(context.Background(), tt.inputID)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, tt.err, err)
			repo.AssertExpectations(t)
		})
	}
}
