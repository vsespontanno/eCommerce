package order

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/order/entity"
	"go.uber.org/zap"
)

// Mocks
type MockPGCartCleaner struct {
	mock.Mock
}

func (m *MockPGCartCleaner) CleanCart(ctx context.Context, order *entity.OrderEvent) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type MockRedisCartCleaner struct {
	mock.Mock
}

func (m *MockRedisCartCleaner) CleanCart(ctx context.Context, order *entity.OrderEvent) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type MockOrderClient struct {
	mock.Mock
}

func (m *MockOrderClient) CreateOrder(ctx context.Context, order *entity.OrderEvent) (string, error) {
	args := m.Called(ctx, order)
	return args.String(0), args.Error(1)
}

func TestCompleteService_CompleteOrder(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockPG := new(MockPGCartCleaner)
		mockRedis := new(MockRedisCartCleaner)
		mockClient := new(MockOrderClient)
		service := NewOrderCompleteService(logger, mockPG, mockRedis, mockClient)

		orderEvent := &entity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
			Status:  "completed",
		}

		mockClient.On("CreateOrder", mock.Anything, orderEvent).Return("order-123", nil)
		mockPG.On("CleanCart", mock.Anything, orderEvent).Return(nil)
		mockRedis.On("CleanCart", mock.Anything, orderEvent).Return(nil)

		err := service.CompleteOrder(context.Background(), orderEvent)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
		mockPG.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("CreateOrder Error", func(t *testing.T) {
		mockPG := new(MockPGCartCleaner)
		mockRedis := new(MockRedisCartCleaner)
		mockClient := new(MockOrderClient)
		service := NewOrderCompleteService(logger, mockPG, mockRedis, mockClient)

		orderEvent := &entity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
		}

		mockClient.On("CreateOrder", mock.Anything, orderEvent).Return("", errors.New("service unavailable"))

		err := service.CompleteOrder(context.Background(), orderEvent)

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
		mockPG.AssertNotCalled(t, "CleanCart", mock.Anything, mock.Anything)
		mockRedis.AssertNotCalled(t, "CleanCart", mock.Anything, mock.Anything)
	})

	t.Run("PGCartCleaner Error", func(t *testing.T) {
		mockPG := new(MockPGCartCleaner)
		mockRedis := new(MockRedisCartCleaner)
		mockClient := new(MockOrderClient)
		service := NewOrderCompleteService(logger, mockPG, mockRedis, mockClient)

		orderEvent := &entity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
		}

		mockClient.On("CreateOrder", mock.Anything, orderEvent).Return("order-123", nil)
		mockPG.On("CleanCart", mock.Anything, orderEvent).Return(errors.New("db error"))

		err := service.CompleteOrder(context.Background(), orderEvent)

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
		mockPG.AssertExpectations(t)
		mockRedis.AssertNotCalled(t, "CleanCart", mock.Anything, mock.Anything)
	})

	t.Run("RedisCartCleaner Error (Non-critical)", func(t *testing.T) {
		mockPG := new(MockPGCartCleaner)
		mockRedis := new(MockRedisCartCleaner)
		mockClient := new(MockOrderClient)
		service := NewOrderCompleteService(logger, mockPG, mockRedis, mockClient)

		orderEvent := &entity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
		}

		mockClient.On("CreateOrder", mock.Anything, orderEvent).Return("order-123", nil)
		mockPG.On("CleanCart", mock.Anything, orderEvent).Return(nil)
		mockRedis.On("CleanCart", mock.Anything, orderEvent).Return(errors.New("redis error"))

		err := service.CompleteOrder(context.Background(), orderEvent)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
		mockPG.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}
