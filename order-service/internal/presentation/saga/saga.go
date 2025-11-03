package saga

import (
	"context"

	"github.com/google/uuid"
	orderEntity "github.com/vsespontanno/eCommerce/order-service/internal/domain/event/entity"
	"github.com/vsespontanno/eCommerce/order-service/internal/domain/product/entity"
	proto "github.com/vsespontanno/eCommerce/proto/order"
	"go.uber.org/zap"
)

type Orchestrator interface {
	SagaTransaction(ctx context.Context, Order orderEntity.OrderEvent) error
}

type SagaServer struct {
	proto.UnimplementedOrderServiceServer
	saga   Orchestrator
	logger *zap.SugaredLogger
}

func NewSagaServer(logger *zap.SugaredLogger) *SagaServer {
	return &SagaServer{logger: logger}
}

func (s *SagaServer) CreateOrderSaga(ctx context.Context, req *proto.CreateOrderSagaRequest) (*proto.CreateOrderSagaResponse, error) {
	var Order orderEntity.OrderEvent
	Order.UserID = req.UserId
	Order.OrderID = uuid.NewString()
	for _, item := range req.Items {
		Order.Products = append(Order.Products, entity.Product{
			ID:       int(item.ProductId),
			Quantity: int(item.Quantity),
		})
	}
	Order.Total = req.Amount
	Order.Status = "Pending"
	err := s.saga.SagaTransaction(ctx, Order)
	if err != nil {
		s.logger.Errorf("error while creating order: %w", err.Error())
		return nil, err
	}
	return &proto.CreateOrderSagaResponse{}, nil
}
