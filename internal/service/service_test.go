package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"wb-l0/internal/domain"

	mocks "wb-l0/internal/service/mocks"
	"wb-l0/pkg/e"
	"wb-l0/pkg/logger"
	"wb-l0/tests"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCaseCreaterOrder(t *testing.T) {
	type mockBehavior func(r *mocks.MockDB, ctx context.Context, o domain.Order)
	testCases := []struct {
		name             string
		inputOrder       domain.Order
		mockBehavior     mockBehavior
		expectedError    error
		expectedResponse string
		expectedID       int
	}{
		{
			name:       "OK",
			inputOrder: tests.InstanceStruct,
			mockBehavior: func(r *mocks.MockDB, ctx context.Context, o domain.Order) {
				r.EXPECT().CreateOrder(ctx, o).Return(1, nil)
				r.EXPECT().GetByUID(ctx, 1).Return(o, nil)
			},
			expectedError:    nil,
			expectedID:       1,
			expectedResponse: tests.InstanceString,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()

			mockOrderService := mocks.NewMockDB(ctrl)

			testCase.mockBehavior(mockOrderService, ctx, testCase.inputOrder)
			logger := logger.SetupPrettySlog()
			service := NewService(logger, mockOrderService, nil)

			id, err := service.CreateOrder(ctx, testCase.inputOrder)
			if testCase.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedID, id)

				order, _ := service.GetByUID(ctx, id)
				actualJSON, _ := service.GetOrderJSON(ctx, order)
				assert.JSONEq(t, testCase.expectedResponse, actualJSON, "JSONs should match")
			}

		})
	}

}

func TestCaseGetOrderByUID(t *testing.T) {
	type mockBehavior func(r *mocks.MockDB, ctx context.Context, id int)

	testCases := []struct {
		name             string
		id               int
		mockBehavior     mockBehavior
		expectedError    error
		expectedResponse domain.Order
	}{
		{
			name:          "OK",
			id:            1,
			expectedError: nil,
			mockBehavior: func(r *mocks.MockDB, ctx context.Context, id int) {
				r.EXPECT().GetByUID(ctx, id).Return(tests.InstanceStruct, nil)
			},
			expectedResponse: tests.InstanceStruct,
		},
		{
			name:          "Not Found",
			id:            2,
			expectedError: fmt.Errorf("service.GetOrder: order not found"),
			mockBehavior: mockBehavior(func(r *mocks.MockDB, ctx context.Context, id int) {
				r.EXPECT().GetByUID(ctx, id).Return(domain.Order{}, e.ErrNotFound)
			}),
			expectedResponse: domain.Order{},
		},
		{
			name:          "InternalError",
			id:            3,
			expectedError: errors.New("service.GetOrder: sql: connection is already closed"),
			mockBehavior: mockBehavior(func(r *mocks.MockDB, ctx context.Context, id int) {
				r.EXPECT().GetByUID(ctx, id).Return(domain.Order{}, sql.ErrConnDone)
			}),
			expectedResponse: domain.Order{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			mockOrderService := mocks.NewMockDB(ctrl)
			testCase.mockBehavior(mockOrderService, ctx, testCase.id)
			logger := logger.SetupPrettySlog()
			service := NewService(logger, mockOrderService, nil)
			order, err := service.GetByUID(ctx, testCase.id)
			if testCase.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, order, testCase.expectedResponse)
			}

		})
	}
}
