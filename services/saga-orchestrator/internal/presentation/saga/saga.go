package saga

import (
	"context"

	"github.com/google/uuid"
	proto "github.com/vsespontanno/eCommerce/proto/saga"
	orderEntity "github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/event/entity"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/domain/product/entity"
	"go.uber.org/zap"
)

type Orchestrator interface {
	SagaTransaction(ctx context.Context, Order orderEntity.OrderEvent) error
}

type SagaServer struct {
	proto.UnimplementedSagaServer
	saga   Orchestrator
	logger *zap.SugaredLogger
}

func NewSagaServer(logger *zap.SugaredLogger, saga Orchestrator) *SagaServer {
	return &SagaServer{logger: logger, saga: saga}
}

func (s *SagaServer) StartCheckout(ctx context.Context, req *proto.StartCheckoutRequest) (*proto.StartCheckoutResponse, error) {
	// Валидация входных данных
	if req.UserID <= 0 {
		s.logger.Errorw("Invalid userID", "userID", req.UserID)
		return &proto.StartCheckoutResponse{OrderID: "", Error: "invalid user ID"}, nil
	}

	if len(req.Cart) == 0 {
		s.logger.Errorw("Empty cart", "userID", req.UserID)
		return &proto.StartCheckoutResponse{OrderID: "", Error: "cart is empty"}, nil
	}

	var Order orderEntity.OrderEvent
	Order.UserID = req.UserID
	Order.OrderID = uuid.NewString()
	Order.Status = "Pending"

	// Формируем заказ и считаем сумму
	for _, item := range req.Cart {
		if item.ProductID <= 0 || item.Quantity <= 0 || item.Price < 0 {
			s.logger.Errorw("Invalid cart item", "productID", item.ProductID, "quantity", item.Quantity, "price", item.Price)
			return &proto.StartCheckoutResponse{OrderID: "", Error: "invalid cart item"}, nil
		}

		Order.Products = append(Order.Products, entity.Product{
			ID:       item.ProductID,
			Quantity: int(item.Quantity),
		})
		Order.Total += item.Price * item.Quantity
	}

	if Order.Total <= 0 {
		s.logger.Errorw("Invalid total amount", "total", Order.Total, "orderID", Order.OrderID)
		return &proto.StartCheckoutResponse{OrderID: "", Error: "invalid total amount"}, nil
	}

	s.logger.Infow("Starting checkout", "orderID", Order.OrderID, "userID", Order.UserID, "total", Order.Total, "items", len(Order.Products))

	err := s.saga.SagaTransaction(ctx, Order)
	if err != nil {
		s.logger.Errorw("Saga transaction failed", "orderID", Order.OrderID, "error", err)
		return &proto.StartCheckoutResponse{OrderID: Order.OrderID, Error: err.Error()}, nil
	}

	s.logger.Infow("Checkout completed successfully", "orderID", Order.OrderID)
	return &proto.StartCheckoutResponse{OrderID: Order.OrderID, Error: ""}, nil
}
