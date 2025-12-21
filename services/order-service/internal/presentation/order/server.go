package ordergrpc

import (
	"context"

	proto "github.com/vsespontanno/eCommerce/proto/orders"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
	"go.uber.org/zap"
)

type OrderSvc interface {
	CreateOrder(ctx context.Context, order *entity.Order) (string, error)
	GetOrder(ctx context.Context, orderID string) (*entity.Order, error)
	ListOrdersByUser(ctx context.Context, userID int64, limit, offset int) ([]entity.Order, error)
}

type Server struct {
	proto.UnimplementedOrderServer
	svc    OrderSvc
	logger *zap.SugaredLogger
}

func NewGRPCServer(svc OrderSvc, logger *zap.SugaredLogger) *Server {
	return &Server{logger: logger, svc: svc}
}

func (s *Server) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	o := req.Order
	if o == nil {
		return &proto.CreateOrderResponse{Error: "empty order"}, nil
	}

	order := entity.Order{
		OrderID: o.OrderId,
		UserID:  o.UserId,
		Total:   o.Total,
		Status:  o.Status,
	}

	// items
	for _, it := range o.Items {
		order.Products = append(order.Products, entity.OrderItem{
			ProductID: it.ProductId,
			Quantity:  it.Quantity,
		})
	}

	id, err := s.svc.CreateOrder(ctx, &order)
	if err != nil {
		s.logger.Errorw("create order failed", "err", err)
		return &proto.CreateOrderResponse{Error: err.Error()}, nil
	}

	return &proto.CreateOrderResponse{OrderId: id}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	if req.OrderId == "" {
		s.logger.Warnw("empty order_id in GetOrder request")
		return &proto.GetOrderResponse{}, nil
	}

	o, err := s.svc.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.logger.Errorw("get order failed", "order_id", req.OrderId, "err", err)
		return nil, err
	}
	if o == nil {
		s.logger.Warnw("order not found", "order_id", req.OrderId)
		return &proto.GetOrderResponse{}, nil
	}

	// Конвертируем entity.Order в proto.OrderEvent
	orderEvent := &proto.OrderEvent{
		OrderId: o.OrderID,
		UserId:  o.UserID,
		Total:   o.Total,
		Status:  o.Status,
	}

	for _, it := range o.Products {
		orderEvent.Items = append(orderEvent.Items, &proto.OrderItem{
			ProductId: it.ProductID,
			Quantity:  it.Quantity,
		})
	}

	return &proto.GetOrderResponse{Order: orderEvent}, nil
}

func (s *Server) ListOrders(ctx context.Context, req *proto.ListOrdersRequest) (*proto.ListOrdersResponse, error) {
	// Валидация
	if req.UserId <= 0 {
		s.logger.Warnw("invalid user_id in ListOrders", "user_id", req.UserId)
		return &proto.ListOrdersResponse{Orders: []*proto.GetOrderResponse{}}, nil
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 10 // default
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	orders, err := s.svc.ListOrdersByUser(ctx, req.UserId, limit, offset)
	if err != nil {
		s.logger.Errorw("list orders failed", "user_id", req.UserId, "err", err)
		return nil, err
	}

	resp := &proto.ListOrdersResponse{
		Orders: make([]*proto.GetOrderResponse, 0, len(orders)),
	}

	for _, o := range orders {
		orderEvent := &proto.OrderEvent{
			OrderId: o.OrderID,
			UserId:  o.UserID,
			Total:   o.Total,
			Status:  o.Status,
		}

		for _, it := range o.Products {
			orderEvent.Items = append(orderEvent.Items, &proto.OrderItem{
				ProductId: it.ProductID,
				Quantity:  it.Quantity,
			})
		}

		resp.Orders = append(resp.Orders, &proto.GetOrderResponse{
			Order: orderEvent,
		})
	}

	s.logger.Infow("orders listed", "user_id", req.UserId, "count", len(orders))
	return resp, nil
}
