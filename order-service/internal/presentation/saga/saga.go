package saga

import (
	"context"

	"github.com/google/uuid"
	orderEntity "github.com/vsespontanno/eCommerce/order-service/internal/domain/event/entity"
	"github.com/vsespontanno/eCommerce/order-service/internal/domain/product/entity"
	proto "github.com/vsespontanno/eCommerce/proto/saga"
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

func (s *SagaServer) CreateOrderSaga(ctx context.Context, req *proto.StartCheckoutRequest) (*proto.StartCheckoutResponse, error) {
	var Order orderEntity.OrderEvent
	Order.UserID = req.UserID
	Order.OrderID = uuid.NewString()
	for _, item := range req.Cart {
		Order.Products = append(Order.Products, entity.Product{
			ID:       int(item.ProductID),
			Quantity: int(item.Quantity),
		})
		Order.Total += item.Price * item.Quantity
	}
	Order.Status = "Pending"
	err := s.saga.SagaTransaction(ctx, Order)
	if err != nil {
		s.logger.Errorf("error while creating order: %w", err.Error())
		return nil, err
	}
	return &proto.StartCheckoutResponse{Success: true, Error: "no error"}, nil
}
