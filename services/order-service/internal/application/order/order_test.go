package order

import (
	"context"
	"errors"
	"testing"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
)

func init() {
	logger.InitLogger()
}

// MockOrderRepo is a mock implementation of interfaces.OrderRepo
type MockOrderRepo struct {
	CreateOrderFunc      func(ctx context.Context, order *entity.Order) error
	GetOrderFunc         func(ctx context.Context, orderID string) (*entity.Order, error)
	ListOrdersByUserFunc func(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error)
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *entity.Order) error {
	return m.CreateOrderFunc(ctx, order)
}

func (m *MockOrderRepo) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	return m.GetOrderFunc(ctx, orderID)
}

func (m *MockOrderRepo) ListOrdersByUser(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
	return m.ListOrdersByUserFunc(ctx, userID, limit, offset)
}

func TestService_CreateOrder(t *testing.T) {
	tests := []struct {
		name          string
		order         *entity.Order
		mockRepo      func() *MockOrderRepo
		expectedID    string
		expectedError error
	}{
		{
			name: "Success",
			order: &entity.Order{
				OrderID: "order-123",
				UserID:  1,
				Total:   1000,
				Status:  "completed",
			},
			mockRepo: func() *MockOrderRepo {
				return &MockOrderRepo{
					CreateOrderFunc: func(ctx context.Context, order *entity.Order) error {
						return nil
					},
				}
			},
			expectedID:    "order-123",
			expectedError: nil,
		},
		{
			name: "Repo Error",
			order: &entity.Order{
				OrderID: "order-123",
			},
			mockRepo: func() *MockOrderRepo {
				return &MockOrderRepo{
					CreateOrderFunc: func(ctx context.Context, order *entity.Order) error {
						return errors.New("db error")
					},
				}
			},
			expectedID:    "",
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOrderService(tt.mockRepo(), logger.Log)

			id, err := service.CreateOrder(context.Background(), tt.order)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if id != tt.expectedID {
					t.Errorf("Expected ID %s, got %s", tt.expectedID, id)
				}
			}
		})
	}
}

func TestService_GetOrder(t *testing.T) {
	tests := []struct {
		name          string
		orderID       string
		mockRepo      func() *MockOrderRepo
		expectedOrder *entity.Order
		expectedError error
	}{
		{
			name:    "Success",
			orderID: "order-123",
			mockRepo: func() *MockOrderRepo {
				return &MockOrderRepo{
					GetOrderFunc: func(ctx context.Context, orderID string) (*entity.Order, error) {
						return &entity.Order{OrderID: "order-123"}, nil
					},
				}
			},
			expectedOrder: &entity.Order{OrderID: "order-123"},
			expectedError: nil,
		},
		{
			name:    "Not Found",
			orderID: "order-123",
			mockRepo: func() *MockOrderRepo {
				return &MockOrderRepo{
					GetOrderFunc: func(ctx context.Context, orderID string) (*entity.Order, error) {
						return nil, errors.New("not found")
					},
				}
			},
			expectedOrder: nil,
			expectedError: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOrderService(tt.mockRepo(), logger.Log)

			order, err := service.GetOrder(context.Background(), tt.orderID)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if order.OrderID != tt.expectedOrder.OrderID {
					t.Errorf("Expected order ID %s, got %s", tt.expectedOrder.OrderID, order.OrderID)
				}
			}
		})
	}
}

func TestService_ListOrdersByUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		limit          uint64
		offset         uint64
		mockRepo       func() *MockOrderRepo
		expectedOrders []entity.Order
		expectedError  error
	}{
		{
			name:   "Success",
			userID: 1,
			limit:  10,
			offset: 0,
			mockRepo: func() *MockOrderRepo {
				return &MockOrderRepo{
					ListOrdersByUserFunc: func(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
						return []entity.Order{{OrderID: "order-1"}, {OrderID: "order-2"}}, nil
					},
				}
			},
			expectedOrders: []entity.Order{{OrderID: "order-1"}, {OrderID: "order-2"}},
			expectedError:  nil,
		},
		{
			name:   "Repo Error",
			userID: 1,
			limit:  10,
			offset: 0,
			mockRepo: func() *MockOrderRepo {
				return &MockOrderRepo{
					ListOrdersByUserFunc: func(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedOrders: nil,
			expectedError:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewOrderService(tt.mockRepo(), logger.Log)

			orders, err := service.ListOrdersByUser(context.Background(), tt.userID, tt.limit, tt.offset)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(orders) != len(tt.expectedOrders) {
					t.Errorf("Expected %d orders, got %d", len(tt.expectedOrders), len(orders))
				}
			}
		})
	}
}
