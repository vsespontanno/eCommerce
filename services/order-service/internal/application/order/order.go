package order

import (
	"context"

	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/interfaces"
	"go.uber.org/zap"
)

type OrderService struct {
	repo   interfaces.OrderRepo
	logger *zap.SugaredLogger
}

func NewOrderService(repo interfaces.OrderRepo, logger *zap.SugaredLogger) *OrderService {
	return &OrderService{repo: repo, logger: logger}
}

// Called by Saga when order is confirmed/reserved
func (s *OrderService) CreateOrder(ctx context.Context, order *entity.Order) (string, error) {
	err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Errorw("repo create order failed", "err", err)
		return "", err
	}

	return order.OrderID, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	return s.repo.GetOrder(ctx, orderID)
}

func (s *OrderService) ListOrdersByUser(ctx context.Context, userID int64, limit, offset int) ([]entity.Order, error) {
	return s.repo.ListOrdersByUser(ctx, userID, limit, offset)
}
