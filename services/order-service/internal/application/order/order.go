package order

import (
	"context"

	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/interfaces"
	"go.uber.org/zap"
)

type Service struct {
	repo   interfaces.OrderRepo
	logger *zap.SugaredLogger
}

func NewOrderService(repo interfaces.OrderRepo, logger *zap.SugaredLogger) *Service {
	return &Service{repo: repo, logger: logger}
}

// Called by Saga when order is confirmed/reserved
func (s *Service) CreateOrder(ctx context.Context, order *entity.Order) (string, error) {
	err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Errorw("repo create order failed", "err", err)
		return "", err
	}

	return order.OrderID, nil
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	return s.repo.GetOrder(ctx, orderID)
}

func (s *Service) ListOrdersByUser(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
	return s.repo.ListOrdersByUser(ctx, userID, limit, offset)
}
