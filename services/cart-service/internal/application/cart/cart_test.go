package cart

import (
	"context"
	"testing"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
)

func init() {
	logger.InitLogger()
}

// MockRedisCartRepo is a mock implementation of RedisCartRepo
type MockRedisCartRepo struct {
	AddNewProductToCartFunc   func(ctx context.Context, userID int64, product *entity.CartItem) error
	SaveCartFunc              func(ctx context.Context, userID int64, cart *entity.Cart) error
	DecrementInCartFunc       func(ctx context.Context, userID int64, productID int64) error
	GetCartFunc               func(ctx context.Context, userID int64) (*entity.Cart, error)
	GetProductFunc            func(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error)
	IncrementInCartFunc       func(ctx context.Context, userID int64, productID int64) error
	RemoveProductFromCartFunc func(ctx context.Context, userID int64, productID int64) error
	DeleteProductFunc         func(ctx context.Context, userID int64, productID int64) error
	ClearCartFunc             func(ctx context.Context, userID int64) error
}

func (m *MockRedisCartRepo) AddNewProductToCart(ctx context.Context, userID int64, product *entity.CartItem) error {
	return m.AddNewProductToCartFunc(ctx, userID, product)
}
func (m *MockRedisCartRepo) SaveCart(ctx context.Context, userID int64, cart *entity.Cart) error {
	return m.SaveCartFunc(ctx, userID, cart)
}
func (m *MockRedisCartRepo) DecrementInCart(ctx context.Context, userID int64, productID int64) error {
	return m.DecrementInCartFunc(ctx, userID, productID)
}
func (m *MockRedisCartRepo) GetCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	return m.GetCartFunc(ctx, userID)
}
func (m *MockRedisCartRepo) GetProduct(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error) {
	return m.GetProductFunc(ctx, userID, productID)
}
func (m *MockRedisCartRepo) IncrementInCart(ctx context.Context, userID int64, productID int64) error {
	return m.IncrementInCartFunc(ctx, userID, productID)
}
func (m *MockRedisCartRepo) RemoveProductFromCart(ctx context.Context, userID int64, productID int64) error {
	return m.RemoveProductFromCartFunc(ctx, userID, productID)
}
func (m *MockRedisCartRepo) DeleteProduct(ctx context.Context, userID int64, productID int64) error {
	return m.DeleteProductFunc(ctx, userID, productID)
}
func (m *MockRedisCartRepo) ClearCart(ctx context.Context, userID int64) error {
	return m.ClearCartFunc(ctx, userID)
}

// MockProducter is a mock implementation of Producter
type MockProducter struct {
	ProductFunc func(ctx context.Context, productID int64) (*entity.CartItem, error)
}

func (m *MockProducter) Product(ctx context.Context, productID int64) (*entity.CartItem, error) {
	return m.ProductFunc(ctx, productID)
}

// MockPostgresCartRepo is a mock implementation of PostgresCartRepo
type MockPostgresCartRepo struct {
	GetCartFunc func(ctx context.Context, userID int64) (*entity.Cart, error)
}

func (m *MockPostgresCartRepo) GetCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	return m.GetCartFunc(ctx, userID)
}

func TestService_Cart(t *testing.T) {
	tests := []struct {
		name         string
		userID       int64
		mockRedis    func() *MockRedisCartRepo
		mockPostgres func() *MockPostgresCartRepo
		expectedCart *entity.Cart
		expectedErr  error
	}{
		{
			name:   "Found in Redis",
			userID: 1,
			mockRedis: func() *MockRedisCartRepo {
				return &MockRedisCartRepo{
					GetCartFunc: func(ctx context.Context, userID int64) (*entity.Cart, error) {
						return &entity.Cart{Items: []entity.CartItem{{UserID: 1}}}, nil
					},
				}
			},
			mockPostgres: func() *MockPostgresCartRepo { return &MockPostgresCartRepo{} },
			expectedCart: &entity.Cart{Items: []entity.CartItem{{UserID: 1}}},
			expectedErr:  nil,
		},
		{
			name:   "Not in Redis, Found in Postgres",
			userID: 1,
			mockRedis: func() *MockRedisCartRepo {
				return &MockRedisCartRepo{
					GetCartFunc: func(ctx context.Context, userID int64) (*entity.Cart, error) {
						return nil, apperrors.ErrNoCartFound
					},
					SaveCartFunc: func(ctx context.Context, userID int64, cart *entity.Cart) error {
						return nil
					},
				}
			},
			mockPostgres: func() *MockPostgresCartRepo {
				return &MockPostgresCartRepo{
					GetCartFunc: func(ctx context.Context, userID int64) (*entity.Cart, error) {
						return &entity.Cart{Items: []entity.CartItem{{UserID: 1}}}, nil
					},
				}
			},
			expectedCart: &entity.Cart{Items: []entity.CartItem{{UserID: 1}}},
			expectedErr:  nil,
		},
		{
			name:   "Not Found Anywhere",
			userID: 1,
			mockRedis: func() *MockRedisCartRepo {
				return &MockRedisCartRepo{
					GetCartFunc: func(ctx context.Context, userID int64) (*entity.Cart, error) {
						return nil, apperrors.ErrNoCartFound
					},
				}
			},
			mockPostgres: func() *MockPostgresCartRepo {
				return &MockPostgresCartRepo{
					GetCartFunc: func(ctx context.Context, userID int64) (*entity.Cart, error) {
						return nil, apperrors.ErrNoCartFound
					},
				}
			},
			expectedCart: &entity.Cart{},
			expectedErr:  apperrors.ErrNoCartFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCart(logger.Log, tt.mockRedis(), nil, tt.mockPostgres(), 10)

			cart, err := service.Cart(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(cart.Items) != len(tt.expectedCart.Items) {
					t.Errorf("Expected %d items, got %d", len(tt.expectedCart.Items), len(cart.Items))
				}
			}
		})
	}
}

func TestService_AddProductToCart(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		productID   int64
		mockRedis   func() *MockRedisCartRepo
		mockProduct func() *MockProducter
		expectedErr error
	}{
		{
			name:      "New Product",
			userID:    1,
			productID: 100,
			mockRedis: func() *MockRedisCartRepo {
				return &MockRedisCartRepo{
					GetProductFunc: func(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error) {
						return nil, apperrors.ErrProductIsNotInCart
					},
					AddNewProductToCartFunc: func(ctx context.Context, userID int64, product *entity.CartItem) error {
						return nil
					},
				}
			},
			mockProduct: func() *MockProducter {
				return &MockProducter{
					ProductFunc: func(ctx context.Context, productID int64) (*entity.CartItem, error) {
						return &entity.CartItem{ProductID: 100}, nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name:      "Existing Product - Increment",
			userID:    1,
			productID: 100,
			mockRedis: func() *MockRedisCartRepo {
				return &MockRedisCartRepo{
					GetProductFunc: func(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error) {
						return &entity.CartItem{ProductID: 100, Quantity: 1}, nil
					},
					IncrementInCartFunc: func(ctx context.Context, userID int64, productID int64) error {
						return nil
					},
				}
			},
			mockProduct: func() *MockProducter { return &MockProducter{} },
			expectedErr: nil,
		},
		{
			name:      "Too Many Products",
			userID:    1,
			productID: 100,
			mockRedis: func() *MockRedisCartRepo {
				return &MockRedisCartRepo{
					GetProductFunc: func(ctx context.Context, userID int64, productID int64) (*entity.CartItem, error) {
						return &entity.CartItem{ProductID: 100, Quantity: 10}, nil
					},
				}
			},
			mockProduct: func() *MockProducter { return &MockProducter{} },
			expectedErr: apperrors.ErrTooManyProductsOfOneType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCart(logger.Log, tt.mockRedis(), tt.mockProduct(), nil, 10)

			err := service.AddProductToCart(context.Background(), tt.userID, tt.productID)

			if tt.expectedErr != nil {
				if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
