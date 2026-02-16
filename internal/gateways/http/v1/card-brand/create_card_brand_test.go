package cardbrand_test

import (
	"context"
	"errors"
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

func TestCreateCardBrandHandler(t *testing.T) {
	ctx := context.Background()
	brandID := uuid.Must(uuid.NewV4())
	now := time.Now()

	tests := []struct {
		name      string
		inputName string
		ucSetup   func(uc *usecases.MockCreateCardBrandUseCase)
		wantID    uuid.UUID
		wantName  string
		wantErr   bool
		errCheck  func(*testing.T, error)
	}{
		{
			name:      "success",
			inputName: "Visa",
			ucSetup: func(uc *usecases.MockCreateCardBrandUseCase) {
				uc.On("CreateCardBrand", mock.Anything, "Visa").Return(&entity.CardBrand{
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
			name:      "returns 409 on duplicate",
			inputName: "Visa",
			ucSetup: func(uc *usecases.MockCreateCardBrandUseCase) {
				uc.On("CreateCardBrand", mock.Anything, "Visa").Return(nil, errs.ErrCardBrandAlreadyExists)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				var statusErr huma.StatusError
				require.True(t, errors.As(err, &statusErr))
				assert.Equal(t, 409, statusErr.GetStatus())
			},
		},
		{
			name:      "returns 500 on generic error",
			inputName: "Visa",
			ucSetup: func(uc *usecases.MockCreateCardBrandUseCase) {
				uc.On("CreateCardBrand", mock.Anything, "Visa").Return(nil, errs.ErrDatabaseGeneric)
			},
			wantErr: true,
			errCheck: func(t *testing.T, err error) {
				var statusErr huma.StatusError
				require.True(t, errors.As(err, &statusErr))
				assert.Equal(t, 500, statusErr.GetStatus())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := usecases.NewMockCreateCardBrandUseCase(t)
			tt.ucSetup(mockUC)

			handler := cardbrand.NewCreateCardBrandHandler(mockUC)

			input := &cardbrand.CreateCardBrandRequest{}
			input.Body.Name = tt.inputName

			result, err := handler.CreateCardBrand(ctx, input)

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
