package usecases_test

import (
	"context"
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

func TestListCardBrandsUC_ListCardBrandsSuccess(t *testing.T) {
	now := time.Now()
	id1 := uuid.Must(uuid.NewV4())
	id2 := uuid.Must(uuid.NewV4())
	brands := []entity.CardBrand{
		{ID: id1, Name: "Visa", CreatedDate: now, LastModifiedDate: now},
		{ID: id2, Name: "Mastercard", CreatedDate: now, LastModifiedDate: now},
	}

	tests := []struct {
		name      string
		filter    ports.ListCardBrandFilterOptions
		repoSetup func(repo *ports.MockListCardBrandRepository)
		want      []entity.CardBrand
	}{
		{
			name:   "returns a list of card brands",
			filter: ports.ListCardBrandFilterOptions{Pagination: page()},
			repoSetup: func(repo *ports.MockListCardBrandRepository) {
				repo.EXPECT().
					ListCardBrands(mock.Anything, ports.ListCardBrandFilterOptions{Pagination: page()}).
					Return(brands, nil)
			},
			want: brands,
		},
		{
			name: "returns empty result",
			filter: ports.ListCardBrandFilterOptions{
				Name:       strPtr("NonExistent"),
				Pagination: page(),
			},
			repoSetup: func(repo *ports.MockListCardBrandRepository) {
				repo.EXPECT().
					ListCardBrands(mock.Anything, ports.ListCardBrandFilterOptions{Name: strPtr("NonExistent"), Pagination: page()}).
					Return([]entity.CardBrand{}, nil)
			},
			want: []entity.CardBrand{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockListCardBrandRepository(t)
			tt.repoSetup(repo)
			uc := usecases.NewListCardBrandUC(repo)
			got, err := uc.ListCardBrands(context.Background(), tt.filter)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestListCardBrandsUC_ListCardBrandsError(t *testing.T) {
	tests := []struct {
		name      string
		filter    ports.ListCardBrandFilterOptions
		repoSetup func(repo *ports.MockListCardBrandRepository)
		err       error
	}{
		{
			name:   "returns error on database generic error",
			filter: ports.ListCardBrandFilterOptions{Pagination: page()},
			repoSetup: func(repo *ports.MockListCardBrandRepository) {
				repo.EXPECT().
					ListCardBrands(mock.Anything, ports.ListCardBrandFilterOptions{Pagination: page()}).
					Return(nil, errs.ErrDatabaseGeneric)
			},
			err: errs.ErrDatabaseGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := ports.NewMockListCardBrandRepository(t)
			tt.repoSetup(repo)
			uc := usecases.NewListCardBrandUC(repo)
			got, err := uc.ListCardBrands(context.Background(), tt.filter)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, tt.err, err)
			repo.AssertExpectations(t)
		})
	}
}

func strPtr(s string) *string { return &s }

func page() domain.Pagination { return domain.Pagination{PageSize: 10, PageNumber: 1} }
