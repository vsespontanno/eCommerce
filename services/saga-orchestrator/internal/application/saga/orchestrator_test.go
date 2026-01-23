package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/config"
	orderEntity "github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/event/entity"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/product/entity"
	"go.uber.org/zap"
)

// Mocks
type MockMoneyReserver struct {
	mock.Mock
}

func (m *MockMoneyReserver) ReserveFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	args := m.Called(ctx, userID, amount)
	return args.String(0), args.Error(1)
}

func (m *MockMoneyReserver) CommitFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	args := m.Called(ctx, userID, amount)
	return args.String(0), args.Error(1)
}

func (m *MockMoneyReserver) ReleaseFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	args := m.Called(ctx, userID, amount)
	return args.String(0), args.Error(1)
}

type MockProductsReserver struct {
	mock.Mock
}

func (m *MockProductsReserver) ReserveProducts(ctx context.Context, productIDs []entity.Product) (bool, error) {
	args := m.Called(ctx, productIDs)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductsReserver) CommitProducts(ctx context.Context, productIDs []entity.Product) (bool, error) {
	args := m.Called(ctx, productIDs)
	return args.Bool(0), args.Error(1)
}

func (m *MockProductsReserver) ReleaseProducts(ctx context.Context, productIDs []entity.Product) (bool, error) {
	args := m.Called(ctx, productIDs)
	return args.Bool(0), args.Error(1)
}

type MockOutboxRepo struct {
	mock.Mock
}

func (m *MockOutboxRepo) SaveEvent(ctx context.Context, event orderEntity.OrderEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestOrchestrator_SagaTransaction(t *testing.T) {
	logger := zap.NewNop().Sugar()
	cfg := &config.Config{}

	t.Run("Success", func(t *testing.T) {
		mockWallet := new(MockMoneyReserver)
		mockProducts := new(MockProductsReserver)
		mockOutbox := new(MockOutboxRepo)
		orchestrator := New(cfg, mockWallet, mockProducts, mockOutbox, logger)

		order := orderEntity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
			Total:   1000,
			Products: []entity.Product{
				{ID: 1, Quantity: 1},
			},
		}

		mockWallet.On("ReserveFunds", mock.Anything, int64(1), int64(1000)).Return("reserved", nil)
		mockProducts.On("ReserveProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("CommitFunds", mock.Anything, int64(1), int64(1000)).Return("committed", nil)
		mockProducts.On("CommitProducts", mock.Anything, order.Products).Return(true, nil)
		mockOutbox.On("SaveEvent", mock.Anything, mock.MatchedBy(func(e orderEntity.OrderEvent) bool {
			return e.Status == "Completed" && e.EventType == orderEntity.EventTypeOrderCompleted
		})).Return(nil)

		err := orchestrator.SagaTransaction(context.Background(), order)

		assert.NoError(t, err)
		mockWallet.AssertExpectations(t)
		mockProducts.AssertExpectations(t)
		mockOutbox.AssertExpectations(t)
	})

	t.Run("Wallet Reserve Failed", func(t *testing.T) {
		mockWallet := new(MockMoneyReserver)
		mockProducts := new(MockProductsReserver)
		mockOutbox := new(MockOutboxRepo)
		orchestrator := New(cfg, mockWallet, mockProducts, mockOutbox, logger)

		order := orderEntity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
			Total:   1000,
		}

		mockWallet.On("ReserveFunds", mock.Anything, int64(1), int64(1000)).Return("", errors.New("insufficient funds"))
		mockWallet.On("ReleaseFunds", mock.Anything, int64(1), int64(1000)).Return("released", nil)

		err := orchestrator.SagaTransaction(context.Background(), order)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet reserve failed")
		mockWallet.AssertExpectations(t)
		mockProducts.AssertNotCalled(t, "ReserveProducts", mock.Anything, mock.Anything)
	})

	t.Run("Products Reserve Failed", func(t *testing.T) {
		mockWallet := new(MockMoneyReserver)
		mockProducts := new(MockProductsReserver)
		mockOutbox := new(MockOutboxRepo)
		orchestrator := New(cfg, mockWallet, mockProducts, mockOutbox, logger)

		order := orderEntity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
			Total:   1000,
			Products: []entity.Product{
				{ID: 1, Quantity: 1},
			},
		}

		mockWallet.On("ReserveFunds", mock.Anything, int64(1), int64(1000)).Return("reserved", nil)
		mockProducts.On("ReserveProducts", mock.Anything, order.Products).Return(false, errors.New("out of stock"))

		// Rollback expectations
		mockProducts.On("ReleaseProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("ReleaseFunds", mock.Anything, int64(1), int64(1000)).Return("released", nil)

		err := orchestrator.SagaTransaction(context.Background(), order)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "products reserve failed")
		mockWallet.AssertExpectations(t)
		mockProducts.AssertExpectations(t)
	})

	t.Run("Wallet Commit Failed", func(t *testing.T) {
		mockWallet := new(MockMoneyReserver)
		mockProducts := new(MockProductsReserver)
		mockOutbox := new(MockOutboxRepo)
		orchestrator := New(cfg, mockWallet, mockProducts, mockOutbox, logger)

		order := orderEntity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
			Total:   1000,
			Products: []entity.Product{
				{ID: 1, Quantity: 1},
			},
		}

		mockWallet.On("ReserveFunds", mock.Anything, int64(1), int64(1000)).Return("reserved", nil)
		mockProducts.On("ReserveProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("CommitFunds", mock.Anything, int64(1), int64(1000)).Return("", errors.New("commit failed"))

		// Rollback expectations
		mockProducts.On("ReleaseProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("ReleaseFunds", mock.Anything, int64(1), int64(1000)).Return("released", nil)

		err := orchestrator.SagaTransaction(context.Background(), order)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wallet commit failed")
		mockWallet.AssertExpectations(t)
		mockProducts.AssertExpectations(t)
	})

	t.Run("Products Commit Failed", func(t *testing.T) {
		mockWallet := new(MockMoneyReserver)
		mockProducts := new(MockProductsReserver)
		mockOutbox := new(MockOutboxRepo)
		orchestrator := New(cfg, mockWallet, mockProducts, mockOutbox, logger)

		order := orderEntity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
			Total:   1000,
			Products: []entity.Product{
				{ID: 1, Quantity: 1},
			},
		}

		mockWallet.On("ReserveFunds", mock.Anything, int64(1), int64(1000)).Return("reserved", nil)
		mockProducts.On("ReserveProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("CommitFunds", mock.Anything, int64(1), int64(1000)).Return("committed", nil)
		mockProducts.On("CommitProducts", mock.Anything, order.Products).Return(false, errors.New("commit failed"))

		// Rollback expectations
		mockProducts.On("ReleaseProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("ReleaseFunds", mock.Anything, int64(1), int64(1000)).Return("released", nil)

		err := orchestrator.SagaTransaction(context.Background(), order)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "products commit failed")
		mockWallet.AssertExpectations(t)
		mockProducts.AssertExpectations(t)
	})

	t.Run("Outbox Save Failed", func(t *testing.T) {
		mockWallet := new(MockMoneyReserver)
		mockProducts := new(MockProductsReserver)
		mockOutbox := new(MockOutboxRepo)
		orchestrator := New(cfg, mockWallet, mockProducts, mockOutbox, logger)

		order := orderEntity.OrderEvent{
			OrderID: "order-123",
			UserID:  1,
			Total:   1000,
			Products: []entity.Product{
				{ID: 1, Quantity: 1},
			},
		}

		mockWallet.On("ReserveFunds", mock.Anything, int64(1), int64(1000)).Return("reserved", nil)
		mockProducts.On("ReserveProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("CommitFunds", mock.Anything, int64(1), int64(1000)).Return("committed", nil)
		mockProducts.On("CommitProducts", mock.Anything, order.Products).Return(true, nil)
		mockOutbox.On("SaveEvent", mock.Anything, mock.Anything).Return(errors.New("db error"))

		// Rollback expectations
		mockProducts.On("ReleaseProducts", mock.Anything, order.Products).Return(true, nil)
		mockWallet.On("ReleaseFunds", mock.Anything, int64(1), int64(1000)).Return("released", nil)

		err := orchestrator.SagaTransaction(context.Background(), order)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save event to outbox")
		mockWallet.AssertExpectations(t)
		mockProducts.AssertExpectations(t)
		mockOutbox.AssertExpectations(t)
	})
}
