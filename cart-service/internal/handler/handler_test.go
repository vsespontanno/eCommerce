package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/cart-service/internal/handler/middleware"
	"go.uber.org/zap"
)

// Mock implementations
type MockCartService struct {
	CartFunc func(ctx context.Context, userID int64) (*models.Cart, error)
}

func (m *MockCartService) Cart(ctx context.Context, userID int64) (*models.Cart, error) {
	if m.CartFunc != nil {
		return m.CartFunc(ctx, userID)
	}
	return &models.Cart{Items: []models.CartItem{}}, nil
}

type MockOrderService struct {
	AddProductToCartFunc    func(ctx context.Context, userID int64, productID int64) error
	GetSelectedProductsFunc func(ctx context.Context, userID int64) (map[int64]int64, error)
	SelectProductFunc       func(ctx context.Context, userID int64, productID int64) error
	UnselectProductFunc     func(ctx context.Context, userID int64, productID int64) error
	ReserveProductsFunc     func(ctx context.Context, userID int64) error
	ReleaseProductsFunc     func(ctx context.Context, userID int64) error
	ConfirmOrderFunc        func(ctx context.Context, userID int64) error
	CancelOrderFunc         func(ctx context.Context, userID int64) error
}

func (m *MockOrderService) AddProductToCart(ctx context.Context, userID int64, productID int64) error {
	if m.AddProductToCartFunc != nil {
		return m.AddProductToCartFunc(ctx, userID, productID)
	}
	return nil
}

func (m *MockOrderService) GetSelectedProducts(ctx context.Context, userID int64) (map[int64]int64, error) {
	if m.GetSelectedProductsFunc != nil {
		return m.GetSelectedProductsFunc(ctx, userID)
	}
	return map[int64]int64{}, nil
}

func (m *MockOrderService) SelectProduct(ctx context.Context, userID int64, productID int64) error {
	if m.SelectProductFunc != nil {
		return m.SelectProductFunc(ctx, userID, productID)
	}
	return nil
}

func (m *MockOrderService) UnselectProduct(ctx context.Context, userID int64, productID int64) error {
	if m.UnselectProductFunc != nil {
		return m.UnselectProductFunc(ctx, userID, productID)
	}
	return nil
}

func (m *MockOrderService) ReserveProducts(ctx context.Context, userID int64) error {
	if m.ReserveProductsFunc != nil {
		return m.ReserveProductsFunc(ctx, userID)
	}
	return nil
}

func (m *MockOrderService) ReleaseProducts(ctx context.Context, userID int64) error {
	if m.ReleaseProductsFunc != nil {
		return m.ReleaseProductsFunc(ctx, userID)
	}
	return nil
}

func (m *MockOrderService) ConfirmOrder(ctx context.Context, userID int64) error {
	if m.ConfirmOrderFunc != nil {
		return m.ConfirmOrderFunc(ctx, userID)
	}
	return nil
}

func (m *MockOrderService) CancelOrder(ctx context.Context, userID int64) error {
	if m.CancelOrderFunc != nil {
		return m.CancelOrderFunc(ctx, userID)
	}
	return nil
}

type MockRateLimiter struct{}

func (m *MockRateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return next
}

type MockValidator struct {
	ValidateTokenFunc func(ctx context.Context, token string) (*models.TokenResponse, error)
}

func (m *MockValidator) ValidateToken(ctx context.Context, token string) (*models.TokenResponse, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx, token)
	}
	return &models.TokenResponse{UserID: 1}, nil
}

// Helper functions
func createTestHandler() (*Handler, *MockCartService, *MockOrderService) {
	logger := zap.NewNop().Sugar()
	mockCartService := &MockCartService{}
	mockOrderService := &MockOrderService{}
	mockValidator := &MockValidator{}

	handler := &Handler{
		cartService:    mockCartService,
		orderService:   mockOrderService,
		sugarLogger:    logger,
		grpcAuthClient: mockValidator,
		rateLimiter:    &MockRateLimiter{},
	}

	return handler, mockCartService, mockOrderService
}

func createTestRequest(method, url string, userID int64) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func createTestRequestWithVars(method, url string, userID int64, vars map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = mux.SetURLVars(req.WithContext(ctx), vars)
	return req
}

// Tests
// Tests
// Tests
func TestHandler_GetCart(t *testing.T) {
	handler, mockCartService, _ := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		cartResponse   *models.Cart
		cartError      error
		expectedStatus int
	}{
		{
			name:   "successful get cart",
			userID: 1,
			cartResponse: &models.Cart{
				Items: []models.CartItem{
					{ID: "1", UserID: 1, ProductID: 100, Quantity: 2},
					{ID: "2", UserID: 1, ProductID: 200, Quantity: 1},
				},
			},
			cartError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty cart",
			userID:         1,
			cartResponse:   &models.Cart{Items: []models.CartItem{}},
			cartError:      nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartService.CartFunc = func(ctx context.Context, userID int64) (*models.Cart, error) {
				return tt.cartResponse, tt.cartError
			}

			req := createTestRequest(http.MethodGet, "/cart", tt.userID)
			w := httptest.NewRecorder()

			handler.GetCart(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_AddProduct(t *testing.T) {
	handler, _, mockOrderService := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		productID      int64
		addError       error
		expectedStatus int
	}{
		{
			name:           "successful add",
			userID:         1,
			productID:      100,
			addError:       nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "too many products",
			userID:         1,
			productID:      100,
			addError:       models.ErrTooManyProductsOfOneType,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "add error",
			userID:         1,
			productID:      100,
			addError:       errors.New("add failed"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderService.AddProductToCartFunc = func(ctx context.Context, userID int64, productID int64) error {
				return tt.addError
			}

			req := createTestRequestWithVars(http.MethodPost, "/cart/100", tt.userID, map[string]string{"id": strconv.FormatInt(tt.productID, 10)})
			w := httptest.NewRecorder()

			handler.AddProduct(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_SelectProduct(t *testing.T) {
	handler, mockCartService, mockOrderService := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		productID      int64
		cartResponse   *models.Cart
		cartError      error
		selectError    error
		expectedStatus int
	}{
		{
			name:      "successful select product",
			userID:    1,
			productID: 100,
			cartResponse: &models.Cart{
				Items: []models.CartItem{
					{ID: "1", UserID: 1, ProductID: 100, Quantity: 1},
				},
			},
			cartError:      nil,
			selectError:    nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:      "product not in wishlist",
			userID:    1,
			productID: 100,
			cartResponse: &models.Cart{
				Items: []models.CartItem{},
			},
			cartError:      nil,
			selectError:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "too many products",
			userID:         1,
			productID:      100,
			cartResponse:   &models.Cart{Items: []models.CartItem{{ID: "1", UserID: 1, ProductID: 100, Quantity: 1}}},
			cartError:      nil,
			selectError:    models.ErrTooManyProductsOfOneType,
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCartService.CartFunc = func(ctx context.Context, userID int64) (*models.Cart, error) {
				return tt.cartResponse, tt.cartError
			}
			mockOrderService.SelectProductFunc = func(ctx context.Context, userID int64, productID int64) error {
				return tt.selectError
			}

			req := createTestRequestWithVars(http.MethodPost, "/cart/selected/100", tt.userID, map[string]string{"id": strconv.FormatInt(tt.productID, 10)})
			w := httptest.NewRecorder()

			handler.SelectProduct(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_UnselectProduct(t *testing.T) {
	handler, _, mockOrderService := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		productID      int64
		unselectError  error
		expectedStatus int
	}{
		{
			name:           "successful unselect",
			userID:         1,
			productID:      100,
			unselectError:  nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unselect error",
			userID:         1,
			productID:      100,
			unselectError:  errors.New("unselect failed"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderService.UnselectProductFunc = func(ctx context.Context, userID int64, productID int64) error {
				return tt.unselectError
			}

			req := createTestRequestWithVars(http.MethodDelete, "/cart/selected/100", tt.userID, map[string]string{"id": strconv.FormatInt(tt.productID, 10)})
			w := httptest.NewRecorder()

			handler.UnselectProduct(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_GetSelected(t *testing.T) {
	handler, _, mockOrderService := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		selectedData   map[int64]int64
		selectedError  error
		expectedStatus int
	}{
		{
			name:   "successful get selected",
			userID: 1,
			selectedData: map[int64]int64{
				100: 2,
				200: 1,
			},
			selectedError:  nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty selected",
			userID:         1,
			selectedData:   map[int64]int64{},
			selectedError:  nil,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderService.GetSelectedProductsFunc = func(ctx context.Context, userID int64) (map[int64]int64, error) {
				return tt.selectedData, tt.selectedError
			}

			req := createTestRequest(http.MethodGet, "/cart/selected", tt.userID)
			w := httptest.NewRecorder()

			handler.GetSelected(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_Checkout(t *testing.T) {
	handler, _, mockOrderService := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		selectedData   map[int64]int64
		selectedError  error
		reserveError   error
		expectedStatus int
	}{
		{
			name:   "successful checkout",
			userID: 1,
			selectedData: map[int64]int64{
				100: 2,
				200: 1,
			},
			selectedError:  nil,
			reserveError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no products selected",
			userID:         1,
			selectedData:   map[int64]int64{},
			selectedError:  nil,
			reserveError:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "reserve products error",
			userID: 1,
			selectedData: map[int64]int64{
				100: 2,
			},
			selectedError:  nil,
			reserveError:   errors.New("reserve failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderService.GetSelectedProductsFunc = func(ctx context.Context, userID int64) (map[int64]int64, error) {
				return tt.selectedData, tt.selectedError
			}
			mockOrderService.ReserveProductsFunc = func(ctx context.Context, userID int64) error {
				return tt.reserveError
			}

			req := createTestRequest(http.MethodPost, "/cart/checkout", tt.userID)
			w := httptest.NewRecorder()

			handler.Checkout(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_ConfirmOrder(t *testing.T) {
	handler, _, mockOrderService := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		confirmError   error
		expectedStatus int
	}{
		{
			name:           "successful confirm",
			userID:         1,
			confirmError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "confirm error",
			userID:         1,
			confirmError:   errors.New("confirm failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderService.ConfirmOrderFunc = func(ctx context.Context, userID int64) error {
				return tt.confirmError
			}

			req := createTestRequest(http.MethodPost, "/cart/confirm", tt.userID)
			w := httptest.NewRecorder()

			handler.ConfirmOrder(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_CancelOrder(t *testing.T) {
	handler, _, mockOrderService := createTestHandler()

	tests := []struct {
		name           string
		userID         int64
		releaseError   error
		cancelError    error
		expectedStatus int
	}{
		{
			name:           "successful cancel",
			userID:         1,
			releaseError:   nil,
			cancelError:    nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "cancel error",
			userID:         1,
			releaseError:   nil,
			cancelError:    errors.New("cancel failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrderService.ReleaseProductsFunc = func(ctx context.Context, userID int64) error {
				return tt.releaseError
			}
			mockOrderService.CancelOrderFunc = func(ctx context.Context, userID int64) error {
				return tt.cancelError
			}

			req := createTestRequest(http.MethodPost, "/cart/cancel", tt.userID)
			w := httptest.NewRecorder()

			handler.CancelOrder(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
