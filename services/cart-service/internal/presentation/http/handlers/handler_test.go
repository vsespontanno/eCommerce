package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/jwt/dto"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/presentation/http/handlers/middleware"
	"go.uber.org/zap"
)

// Mocks
type MockCartService struct {
	mock.Mock
}

func (m *MockCartService) Cart(ctx context.Context, userID int64) (*entity.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

func (m *MockCartService) AddProductToCart(ctx context.Context, userID int64, productID int64) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func (m *MockCartService) Increment(ctx context.Context, userID int64, productID int64) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func (m *MockCartService) Decrement(ctx context.Context, userID int64, productID int64) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

func (m *MockCartService) ClearCart(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockCartService) DeleteProductFromCart(ctx context.Context, userID int64, productID int64) error {
	args := m.Called(ctx, userID, productID)
	return args.Error(0)
}

type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return next
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) ValidateToken(ctx context.Context, token string) (*dto.TokenResponse, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenResponse), args.Error(1)
}

type MockCheckouter struct {
	mock.Mock
}

func (m *MockCheckouter) Checkout(ctx context.Context, userID int64) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func TestHandler_GetCart(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		cart := &entity.Cart{
			Items: []entity.CartItem{
				{ProductID: 1, Quantity: 2, Price: 100},
			},
		}

		mockService.On("Cart", mock.Anything, int64(1)).Return(cart, nil)

		req := httptest.NewRequest(http.MethodGet, "/cart", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		w := httptest.NewRecorder()

		handler.GetCart(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Empty Cart", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("Cart", mock.Anything, int64(1)).Return(nil, apperrors.ErrNoCartFound)

		req := httptest.NewRequest(http.MethodGet, "/cart", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		w := httptest.NewRecorder()

		handler.GetCart(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Your cart is empty", response["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("Cart", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		req := httptest.NewRequest(http.MethodGet, "/cart", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		w := httptest.NewRecorder()

		handler.GetCart(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_ClearCart(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("ClearCart", mock.Anything, int64(1)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/cart", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		w := httptest.NewRecorder()

		handler.ClearCart(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("ClearCart", mock.Anything, int64(1)).Return(errors.New("db error"))

		req := httptest.NewRequest(http.MethodDelete, "/cart", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		w := httptest.NewRecorder()

		handler.ClearCart(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_RemoveProduct(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("DeleteProductFromCart", mock.Anything, int64(1), int64(100)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/cart/100", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req = mux.SetURLVars(req, map[string]string{"id": "100"})
		w := httptest.NewRecorder()

		handler.RemoveProduct(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid ID", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/cart/invalid", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
		w := httptest.NewRecorder()

		handler.RemoveProduct(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Error", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("DeleteProductFromCart", mock.Anything, int64(1), int64(100)).Return(errors.New("db error"))

		req := httptest.NewRequest(http.MethodDelete, "/cart/100", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req = mux.SetURLVars(req, map[string]string{"id": "100"})
		w := httptest.NewRecorder()

		handler.RemoveProduct(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_IncrementProduct(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("AddProductToCart", mock.Anything, int64(1), int64(100)).Return(nil)

		req := httptest.NewRequest(http.MethodPatch, "/cart/100/increment", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req = mux.SetURLVars(req, map[string]string{"id": "100"})
		w := httptest.NewRecorder()

		handler.IncrementProduct(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Limit Exceeded", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("AddProductToCart", mock.Anything, int64(1), int64(100)).Return(apperrors.ErrTooManyProductsOfOneType)

		req := httptest.NewRequest(http.MethodPatch, "/cart/100/increment", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req = mux.SetURLVars(req, map[string]string{"id": "100"})
		w := httptest.NewRecorder()

		handler.IncrementProduct(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_DecrementProduct(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCartService)
		handler := New(mockService, logger, nil, nil, nil)

		mockService.On("Decrement", mock.Anything, int64(1), int64(100)).Return(nil)

		req := httptest.NewRequest(http.MethodPatch, "/cart/100/decrement", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		req = mux.SetURLVars(req, map[string]string{"id": "100"})
		w := httptest.NewRecorder()

		handler.DecrementProduct(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_Checkout(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockCheckouter := new(MockCheckouter)
		handler := New(nil, logger, nil, nil, mockCheckouter)

		mockCheckouter.On("Checkout", mock.Anything, int64(1)).Return("order-123", nil)

		req := httptest.NewRequest(http.MethodPost, "/cart/order/checkout", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		w := httptest.NewRecorder()

		handler.Checkout(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "order-123", response["orderId"])
		mockCheckouter.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockCheckouter := new(MockCheckouter)
		handler := New(nil, logger, nil, nil, mockCheckouter)

		mockCheckouter.On("Checkout", mock.Anything, int64(1)).Return("", errors.New("checkout error"))

		req := httptest.NewRequest(http.MethodPost, "/cart/order/checkout", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, int64(1)))
		w := httptest.NewRecorder()

		handler.Checkout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockCheckouter.AssertExpectations(t)
	})
}

func TestHandler_HealthCheck(t *testing.T) {
	logger := zap.NewNop().Sugar()
	handler := New(nil, logger, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "healthy", response["status"])
}
