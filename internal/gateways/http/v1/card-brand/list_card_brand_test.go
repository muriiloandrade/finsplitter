package cardbrand_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/muriiloandrade/finsplitter/internal/app/ports"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/card-brand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListCardBrandsHandler(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	emptyStr := ""

	tests := []struct {
		name     string
		input    *cardbrand.ListCardBrandsRequest
		ucSetup  func(uc *usecases.MockListCardBrandsUseCase)
		want     []entity.CardBrand
		wantErr  bool
		errCheck func(*testing.T, error)
	}{
		{
			name: "success",
			input: &cardbrand.ListCardBrandsRequest{
				PageSize:   10,
				PageNumber: 1,
			},
			ucSetup: func(uc *usecases.MockListCardBrandsUseCase) {
				brands := []entity.CardBrand{
					{ID: uuid.Must(uuid.NewV4()), Name: "Visa", CreatedDate: now, LastModifiedDate: now},
					{ID: uuid.Must(uuid.NewV4()), Name: "Mastercard", CreatedDate: now, LastModifiedDate: now},
				}
				uc.On("ListCardBrands", mock.Anything, ports.ListCardBrandFilterOptions{
					PageSize: 10, PageNumber: 1, Name: &emptyStr,
				}).Return(brands, nil)
			},
			want: []entity.CardBrand{
				{Name: "Visa"},
				{Name: "Mastercard"},
			},
			wantErr: false,
		},
		{
			name: "returns empty list",
			input: &cardbrand.ListCardBrandsRequest{
				PageSize:   10,
				PageNumber: 1,
			},
			ucSetup: func(uc *usecases.MockListCardBrandsUseCase) {
				uc.On("ListCardBrands", mock.Anything, ports.ListCardBrandFilterOptions{
					PageSize: 10, PageNumber: 1, Name: &emptyStr,
				}).Return([]entity.CardBrand{}, nil)
			},
			want:    []entity.CardBrand{},
			wantErr: false,
		},
		{
			name: "returns 500 on error",
			input: &cardbrand.ListCardBrandsRequest{
				PageSize:   10,
				PageNumber: 1,
			},
			ucSetup: func(uc *usecases.MockListCardBrandsUseCase) {
				uc.On("ListCardBrands", mock.Anything, ports.ListCardBrandFilterOptions{
					PageSize: 10, PageNumber: 1, Name: &emptyStr,
				}).Return(nil, errs.ErrDatabaseGeneric)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				var statusErr huma.StatusError
				require.True(t, errors.As(err, &statusErr))
				assert.Equal(t, 500, statusErr.GetStatus())
			},
		},
		{
			name: "filters by name",
			input: &cardbrand.ListCardBrandsRequest{
				Name:       "Visa",
				PageSize:   10,
				PageNumber: 1,
			},
			ucSetup: func(uc *usecases.MockListCardBrandsUseCase) {
				name := "Visa"
				brands := []entity.CardBrand{
					{ID: uuid.Must(uuid.NewV4()), Name: "Visa", CreatedDate: now, LastModifiedDate: now},
				}
				uc.On("ListCardBrands", mock.Anything, ports.ListCardBrandFilterOptions{
					Name: &name, PageSize: 10, PageNumber: 1,
				}).Return(brands, nil)
			},
			want:    []entity.CardBrand{{Name: "Visa"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := usecases.NewMockListCardBrandsUseCase(t)
			tt.ucSetup(mockUC)

			handler := cardbrand.NewListCardBrandsHandler(mockUC)

			result, err := handler.ListCardBrands(ctx, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, result.Body.CardBrands, len(tt.want))
			}

			mockUC.AssertExpectations(t)
		})
	}
}
