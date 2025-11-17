package order

import (
	"context"

	"github.com/vsespontanno/eCommerce/order-service/internal/domain/order/entity"
	"github.com/vsespontanno/eCommerce/order-service/internal/domain/order/interfaces"
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
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem, total int64) (string, error) {
	order := &entity.Order{
		UserID: userID,
		Total:  total,
		Status: "pending",
		Items:  items,
	}
	id, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		s.logger.Errorw("repo create order failed", "err", err)
		return "", err
	}
	order.ID = id

	return id, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	return s.repo.GetOrder(ctx, orderID)
}

func (s *OrderService) ListOrdersByUser(ctx context.Context, userID int64, limit, offset int) ([]entity.Order, error) {
	return s.repo.ListOrdersByUser(ctx, userID, limit, offset)
}
