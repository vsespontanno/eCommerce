package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	proto "github.com/vsespontanno/eCommerce/proto/saga"
	orderEntity "github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/event/entity"
	"go.uber.org/zap"
)

// Mocks
type MockOrchestrator struct {
	mock.Mock
}

func (m *MockOrchestrator) SagaTransaction(ctx context.Context, Order orderEntity.OrderEvent) error {
	args := m.Called(ctx, Order)
	return args.Error(0)
}

func TestServer_StartCheckout(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("Success", func(t *testing.T) {
		mockOrchestrator := new(MockOrchestrator)
		server := NewSagaServer(logger, mockOrchestrator)

		req := &proto.StartCheckoutRequest{
			UserID: 1,
			Cart: []*proto.Cart{
				{ProductID: 1, Quantity: 1, Price: 100},
			},
		}

		mockOrchestrator.On("SagaTransaction", mock.Anything, mock.MatchedBy(func(order orderEntity.OrderEvent) bool {
			return order.UserID == 1 && order.Total == 100 && len(order.Products) == 1
		})).Return(nil)

		resp, err := server.StartCheckout(context.Background(), req)

		assert.NoError(t, err)
		assert.NotEmpty(t, resp.OrderID)
		assert.Empty(t, resp.Error)
		mockOrchestrator.AssertExpectations(t)
	})

	t.Run("Invalid UserID", func(t *testing.T) {
		mockOrchestrator := new(MockOrchestrator)
		server := NewSagaServer(logger, mockOrchestrator)

		req := &proto.StartCheckoutRequest{
			UserID: 0,
			Cart: []*proto.Cart{
				{ProductID: 1, Quantity: 1, Price: 100},
			},
		}

		resp, err := server.StartCheckout(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, resp.OrderID)
		assert.Equal(t, "invalid user ID", resp.Error)
		mockOrchestrator.AssertNotCalled(t, "SagaTransaction", mock.Anything, mock.Anything)
	})

	t.Run("Empty Cart", func(t *testing.T) {
		mockOrchestrator := new(MockOrchestrator)
		server := NewSagaServer(logger, mockOrchestrator)

		req := &proto.StartCheckoutRequest{
			UserID: 1,
			Cart:   []*proto.Cart{},
		}

		resp, err := server.StartCheckout(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, resp.OrderID)
		assert.Equal(t, "cart is empty", resp.Error)
		mockOrchestrator.AssertNotCalled(t, "SagaTransaction", mock.Anything, mock.Anything)
	})

	t.Run("Invalid Cart Item", func(t *testing.T) {
		mockOrchestrator := new(MockOrchestrator)
		server := NewSagaServer(logger, mockOrchestrator)

		req := &proto.StartCheckoutRequest{
			UserID: 1,
			Cart: []*proto.Cart{
				{ProductID: 0, Quantity: 1, Price: 100},
			},
		}

		resp, err := server.StartCheckout(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, resp.OrderID)
		assert.Equal(t, "invalid cart item", resp.Error)
		mockOrchestrator.AssertNotCalled(t, "SagaTransaction", mock.Anything, mock.Anything)
	})

	t.Run("Invalid Total Amount", func(t *testing.T) {
		mockOrchestrator := new(MockOrchestrator)
		server := NewSagaServer(logger, mockOrchestrator)

		req := &proto.StartCheckoutRequest{
			UserID: 1,
			Cart: []*proto.Cart{
				{ProductID: 1, Quantity: 1, Price: 0},
			},
		}

		resp, err := server.StartCheckout(context.Background(), req)

		assert.NoError(t, err)
		assert.Empty(t, resp.OrderID)
		assert.Equal(t, "invalid total amount", resp.Error)
		mockOrchestrator.AssertNotCalled(t, "SagaTransaction", mock.Anything, mock.Anything)
	})

	t.Run("Saga Transaction Failed", func(t *testing.T) {
		mockOrchestrator := new(MockOrchestrator)
		server := NewSagaServer(logger, mockOrchestrator)

		req := &proto.StartCheckoutRequest{
			UserID: 1,
			Cart: []*proto.Cart{
				{ProductID: 1, Quantity: 1, Price: 100},
			},
		}

		mockOrchestrator.On("SagaTransaction", mock.Anything, mock.Anything).Return(errors.New("saga error"))

		resp, err := server.StartCheckout(context.Background(), req)

		assert.NoError(t, err)
		assert.NotEmpty(t, resp.OrderID)
		assert.Equal(t, "saga error", resp.Error)
		mockOrchestrator.AssertExpectations(t)
	})
}
