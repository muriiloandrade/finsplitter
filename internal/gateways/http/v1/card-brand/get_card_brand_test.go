package cardbrand_test

import (
	"context"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	usecases "github.com/muriiloandrade/finsplitter/internal/app/usecases/card-brand"
	"github.com/muriiloandrade/finsplitter/internal/domain/entity"
	"github.com/muriiloandrade/finsplitter/internal/domain/errs"
	"github.com/muriiloandrade/finsplitter/internal/gateways/http/v1/card-brand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetCardBrandHandler(t *testing.T) {
	ctx := context.Background()
	brandID := uuid.Must(uuid.NewV4())
	now := time.Now()

	tests := []struct {
		name     string
		inputID  uuid.UUID
		ucSetup  func(uc *usecases.MockGetCardBrandByIDUseCase)
		wantID   uuid.UUID
		wantName string
		wantErr  bool
		errCheck func(*testing.T, error)
	}{
		{
			name:    "success",
			inputID: brandID,
			ucSetup: func(uc *usecases.MockGetCardBrandByIDUseCase) {
				uc.On("GetCardBrandByID", mock.Anything, brandID).Return(&entity.CardBrand{
					ID:               brandID,
					Name:             "Visa",
					CreatedDate:      now,
					LastModifiedDate: now,
				}, nil)
			},
			wantID:   brandID,
			wantName: "Visa",
			wantErr:  false,
		},
		{
			name:    "returns 404 on not found",
			inputID: brandID,
			ucSetup: func(uc *usecases.MockGetCardBrandByIDUseCase) {
				uc.On("GetCardBrandByID", mock.Anything, brandID).Return(nil, errs.ErrCardBrandNotFound)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				var statusErr huma.StatusError
				require.ErrorAs(t, err, &statusErr)
				assert.Equal(t, 404, statusErr.GetStatus())
			},
		},
		{
			name:    "returns 500 on generic error",
			inputID: brandID,
			ucSetup: func(uc *usecases.MockGetCardBrandByIDUseCase) {
				uc.On("GetCardBrandByID", mock.Anything, brandID).Return(nil, errs.ErrDatabaseGeneric)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				var statusErr huma.StatusError
				require.ErrorAs(t, err, &statusErr)
				assert.Equal(t, 500, statusErr.GetStatus())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := usecases.NewMockGetCardBrandByIDUseCase(t)
			tt.ucSetup(mockUC)

			handler := cardbrand.NewGetCardBrandHandler(mockUC)

			input := &cardbrand.GetCardBrandRequest{ID: tt.inputID}

			result, err := handler.GetCardBrand(ctx, input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCheck != nil {
					tt.errCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantID, result.Body.ID)
				assert.Equal(t, tt.wantName, result.Body.Name)
			}

			mockUC.AssertExpectations(t)
		})
	}
}
