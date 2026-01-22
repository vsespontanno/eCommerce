package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"go.uber.org/zap"
)

// Mocks
type MockCarter struct {
	mock.Mock
}

func (m *MockCarter) GetCartProducts(ctx context.Context, userID int64) (*entity.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Cart), args.Error(1)
}

type MockSagaClient struct {
	mock.Mock
}

func (m *MockSagaClient) StartCheckout(ctx context.Context, userID int64, cart *entity.Cart) (string, error) {
	args := m.Called(ctx, userID, cart)
	return args.String(0), args.Error(1)
}

func TestService_Checkout(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockCarter := new(MockCarter)
		mockSaga := new(MockSagaClient)
		service := NewSagaService(logger, mockCarter, mockSaga)

		cart := &entity.Cart{
			Items: []entity.CartItem{
				{ProductID: 1, Quantity: 1, Price: 100},
			},
		}

		mockCarter.On("GetCartProducts", mock.Anything, int64(1)).Return(cart, nil)
		mockSaga.On("StartCheckout", mock.Anything, int64(1), cart).Return("order-123", nil)

		orderID, err := service.Checkout(context.Background(), 1)

		assert.NoError(t, err)
		assert.Equal(t, "order-123", orderID)
		mockCarter.AssertExpectations(t)
		mockSaga.AssertExpectations(t)
	})

	t.Run("GetCartProducts Error", func(t *testing.T) {
		mockCarter := new(MockCarter)
		mockSaga := new(MockSagaClient)
		service := NewSagaService(logger, mockCarter, mockSaga)

		mockCarter.On("GetCartProducts", mock.Anything, int64(1)).Return(nil, errors.New("redis error"))

		orderID, err := service.Checkout(context.Background(), 1)

		assert.Error(t, err)
		assert.Empty(t, orderID)
		mockCarter.AssertExpectations(t)
		mockSaga.AssertNotCalled(t, "StartCheckout", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("StartCheckout Error", func(t *testing.T) {
		mockCarter := new(MockCarter)
		mockSaga := new(MockSagaClient)
		service := NewSagaService(logger, mockCarter, mockSaga)

		cart := &entity.Cart{
			Items: []entity.CartItem{
				{ProductID: 1, Quantity: 1, Price: 100},
			},
		}

		mockCarter.On("GetCartProducts", mock.Anything, int64(1)).Return(cart, nil)
		mockSaga.On("StartCheckout", mock.Anything, int64(1), cart).Return("", errors.New("saga error"))

		orderID, err := service.Checkout(context.Background(), 1)

		assert.Error(t, err)
		assert.Empty(t, orderID)
		mockCarter.AssertExpectations(t)
		mockSaga.AssertExpectations(t)
	})
}
