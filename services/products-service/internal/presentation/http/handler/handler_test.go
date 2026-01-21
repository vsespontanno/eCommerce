package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/products/entity"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/http/handler/middleware"
)

func init() {
	logger.InitLogger()
}

// MockProductStorer is a mock implementation of ProductStorer
type MockProductStorer struct {
	GetProductsFunc    func(ctx context.Context) ([]*entity.Product, error)
	GetProductByIDFunc func(ctx context.Context, id int64) (*entity.Product, error)
}

func (m *MockProductStorer) SaveProduct(ctx context.Context, product *entity.Product) error {
	return nil
}
func (m *MockProductStorer) GetProducts(ctx context.Context) ([]*entity.Product, error) {
	return m.GetProductsFunc(ctx)
}
func (m *MockProductStorer) GetProductByID(ctx context.Context, id int64) (*entity.Product, error) {
	return m.GetProductByIDFunc(ctx, id)
}
func (m *MockProductStorer) GetProductsByID(ctx context.Context, ids []int64) ([]*entity.Product, error) {
	return nil, nil
}

// MockCartStorer is a mock implementation of CartStorer
type MockCartStorer struct {
	UpsertProductToCartFunc func(ctx context.Context, userID int64, productID int64, amountForProduct int64) (int, error)
}

func (m *MockCartStorer) UpsertProductToCart(ctx context.Context, userID int64, productID int64, amountForProduct int64) (int, error) {
	return m.UpsertProductToCartFunc(ctx, userID, productID, amountForProduct)
}

func TestHandler_GetProducts(t *testing.T) {
	tests := []struct {
		name           string
		mockStore      func() *MockProductStorer
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "Success",
			mockStore: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductsFunc: func(ctx context.Context) ([]*entity.Product, error) {
						return []*entity.Product{
							{ID: 1, Name: "Product 1"},
							{ID: 2, Name: "Product 2"},
						}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "Internal Error",
			mockStore: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductsFunc: func(ctx context.Context) ([]*entity.Product, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(nil, tt.mockStore(), logger.Log, nil)

			req, _ := http.NewRequestWithContext(context.Background(), "GET", "/products", nil)
			rr := httptest.NewRecorder()

			h.GetProducts(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var products []*entity.Product
				json.NewDecoder(rr.Body).Decode(&products)
				if len(products) != tt.expectedCount {
					t.Errorf("Expected %d products, got %d", tt.expectedCount, len(products))
				}
			}
		})
	}
}

func TestHandler_GetProduct(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockStore      func() *MockProductStorer
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			mockStore: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return &entity.Product{ID: 1, Name: "Product 1"}, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   "1",
			mockStore: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return nil, apperrors.ErrNoProductFound
					},
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Invalid ID",
			id:   "abc",
			mockStore: func() *MockProductStorer {
				return &MockProductStorer{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Internal Error",
			id:   "1",
			mockStore: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(nil, tt.mockStore(), logger.Log, nil)

			req, _ := http.NewRequestWithContext(context.Background(), "GET", "/products/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			rr := httptest.NewRecorder()

			h.GetProduct(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestHandler_AddProductToCart(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		userID         any
		mockProduct    func() *MockProductStorer
		mockCart       func() *MockCartStorer
		expectedStatus int
	}{
		{
			name:   "Success",
			id:     "1",
			userID: int64(1),
			mockProduct: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return &entity.Product{ID: 1, Price: 100}, nil
					},
				}
			},
			mockCart: func() *MockCartStorer {
				return &MockCartStorer{
					UpsertProductToCartFunc: func(ctx context.Context, userID int64, productID int64, amountForProduct int64) (int, error) {
						return 1, nil
					},
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized",
			id:             "1",
			userID:         nil,
			mockProduct:    func() *MockProductStorer { return &MockProductStorer{} },
			mockCart:       func() *MockCartStorer { return &MockCartStorer{} },
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Product ID",
			id:             "abc",
			userID:         int64(1),
			mockProduct:    func() *MockProductStorer { return &MockProductStorer{} },
			mockCart:       func() *MockCartStorer { return &MockCartStorer{} },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Product Not Found",
			id:     "1",
			userID: int64(1),
			mockProduct: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return nil, apperrors.ErrNoProductFound
					},
				}
			},
			mockCart:       func() *MockCartStorer { return &MockCartStorer{} },
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Cart Error",
			id:     "1",
			userID: int64(1),
			mockProduct: func() *MockProductStorer {
				return &MockProductStorer{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return &entity.Product{ID: 1, Price: 100}, nil
					},
				}
			},
			mockCart: func() *MockCartStorer {
				return &MockCartStorer{
					UpsertProductToCartFunc: func(ctx context.Context, userID int64, productID int64, amountForProduct int64) (int, error) {
						return 0, errors.New("cart error")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(tt.mockCart(), tt.mockProduct(), logger.Log, nil)

			req, _ := http.NewRequestWithContext(context.Background(), "POST", "/products/"+tt.id+"/add-to-cart", nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			h.AddProductToCart(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
