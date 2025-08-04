package rest

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"wb-l0/internal/domain"
	handler_mocks "wb-l0/internal/ports/rest/mocks"
	"wb-l0/pkg/e"
	"wb-l0/pkg/logger"
	"wb-l0/tests"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_CreateOrderHandler(t *testing.T) {
	type mockBehavior func(r *handler_mocks.MockOrderStorage, ctx context.Context, o domain.Order)
	testCases := []struct {
		name               string
		inputOrder         domain.Order
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:             "OK",
			inputOrder:       tests.InstanceStruct,
			expectedResponse: `{"The order successfully created, id":1}`,
			mockBehavior: func(r *handler_mocks.MockOrderStorage, ctx context.Context, o domain.Order) {
				r.EXPECT().CreateOrder(gomock.Any(), o).Return(1, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Error",
			inputOrder: tests.InstanceStruct,
			mockBehavior: func(r *handler_mocks.MockOrderStorage, ctx context.Context, o domain.Order) {
				r.EXPECT().CreateOrder(gomock.Any(), o).Return(1, nil)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"The order successfully created, id":1}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var ctrl *gomock.Controller
			if testCase.mockBehavior != nil {
				ctrl = gomock.NewController(t)
				defer ctrl.Finish()
			}
			ctx := context.Background()

			mockOrderHandler := handler_mocks.NewMockOrderStorage(ctrl)
			mockRenderHandler := handler_mocks.NewMockServiceRender(ctrl)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockOrderHandler, ctx, testCase.inputOrder)
			}

			logger := logger.SetupPrettySlog()
			handler := NewHandler(logger, mockOrderHandler, mockRenderHandler)

			r := gin.Default()
			r.POST("/order", handler.PostHandler)

			json, err := json.Marshal(testCase.inputOrder)
			if err != nil {
				log.Println("failed to marshal inputOrder")
				return
			}

			w := httptest.NewRecorder()

			req := httptest.NewRequest("POST", "/order", bytes.NewBufferString(string(json)))
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			if testCase.expectedStatusCode == http.StatusOK {
				assert.JSONEq(t, testCase.expectedResponse, w.Body.String())
			} else {
				assert.Equal(t, testCase.expectedResponse, w.Body.String())

			}
		})
	}
}

func Test_GetByUIDHandler(t *testing.T) {
	type mockBehavior func(r *handler_mocks.MockOrderStorage, ctx context.Context, id int)
	testCases := []struct {
		name               string
		id                 int
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedResponse   string
		requestURL         string
	}{
		{
			name: "OK",
			id:   1,
			mockBehavior: mockBehavior(func(r *handler_mocks.MockOrderStorage, ctx context.Context, id int) {
				r.EXPECT().GetByUID(gomock.Any(), id).Return(tests.InstanceStruct, nil)
			}),
			expectedResponse:   fmt.Sprintf(`{"Response": %s}`, tests.InstanceString),
			expectedStatusCode: http.StatusOK,
			requestURL:         "/orders/1",
		},
		{
			name: "Not Found",
			id:   2,
			mockBehavior: mockBehavior(func(r *handler_mocks.MockOrderStorage, ctx context.Context, id int) {
				r.EXPECT().GetByUID(gomock.Any(), id).Return(domain.Order{}, e.ErrNotFound)
			}),
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   fmt.Sprintf(`{"error":"%s"}`, e.ErrNotFound.Error()),
			requestURL:         "/orders/2",
		},
		{
			name: "InternalError",
			id:   3,
			mockBehavior: func(r *handler_mocks.MockOrderStorage, ctx context.Context, id int) {
				r.EXPECT().GetByUID(gomock.Any(), id).Return(domain.Order{}, sql.ErrConnDone)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   `{"error":"failed to perform func GetByUID"}`,
			requestURL:         "/orders/3",
		},

		{
			name:               "Invalid id - not numeric",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"error":"invalid id"}`,
			requestURL:         "/orders/abc",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var ctrl *gomock.Controller
			if testCase.mockBehavior != nil {
				ctrl = gomock.NewController(t)
				defer ctrl.Finish()
			}
			ctx := context.Background()

			mockOrderHandler := handler_mocks.NewMockOrderStorage(ctrl)
			mockRenderHandler := handler_mocks.NewMockServiceRender(ctrl)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mockOrderHandler, ctx, testCase.id)
			}

			logger := logger.SetupPrettySlog()
			handler := NewHandler(logger, mockOrderHandler, mockRenderHandler)

			r := gin.Default()
			r.GET("/orders/:id", handler.GetOrder)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", testCase.requestURL, nil)
			req.Header.Set("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			if testCase.expectedStatusCode == http.StatusOK {
				assert.JSONEq(t, testCase.expectedResponse, w.Body.String())
			} else {
				assert.Equal(t, testCase.expectedResponse, w.Body.String())

			}

		})
	}
}
