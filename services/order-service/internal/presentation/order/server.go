package ordergrpc

import (
	"context"

	"github.com/google/uuid"
	proto "github.com/vsespontanno/eCommerce/proto/orders"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderSvc interface {
	CreateOrder(ctx context.Context, order *entity.Order) (string, error)
	GetOrder(ctx context.Context, orderID string) (*entity.Order, error)
	ListOrdersByUser(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error)
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
		s.logger.Warnw("CreateOrder called with empty order")
		return nil, status.Error(codes.InvalidArgument, "order is required")
	}

	if o.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	if _, err := uuid.Parse(o.OrderId); err != nil {
		s.logger.Warnw("Invalid order_id format", "order_id", o.OrderId)
		return nil, status.Error(codes.InvalidArgument, "order_id must be a valid UUID")
	}

	if o.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}

	order := entity.Order{
		OrderID: o.OrderId,
		UserID:  o.UserId,
		Total:   o.Total,
		Status:  o.Status,
	}

	// Convert items
	for _, it := range o.Items {
		if it.ProductId <= 0 || it.Quantity <= 0 {
			s.logger.Warnw("Invalid order item", "product_id", it.ProductId, "quantity", it.Quantity)
			continue
		}
		order.Products = append(order.Products, entity.OrderItem{
			ProductID: it.ProductId,
			Quantity:  it.Quantity,
		})
	}

	id, err := s.svc.CreateOrder(ctx, &order)
	if err != nil {
		s.logger.Errorw("create order failed", "order_id", o.OrderId, "err", err)
		return nil, status.Error(codes.Internal, "failed to create order")
	}

	return &proto.CreateOrderResponse{OrderId: id}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	if req.OrderId == "" {
		s.logger.Warnw("empty order_id in GetOrder request")
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	// Validate UUID format
	if _, err := uuid.Parse(req.OrderId); err != nil {
		s.logger.Warnw("Invalid order_id format in GetOrder", "order_id", req.OrderId)
		return nil, status.Error(codes.InvalidArgument, "order_id must be a valid UUID")
	}

	o, err := s.svc.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.logger.Errorw("get order failed", "order_id", req.OrderId, "err", err)
		return nil, status.Error(codes.Internal, "failed to get order")
	}
	if o == nil {
		s.logger.Warnw("order not found", "order_id", req.OrderId)
		return nil, status.Error(codes.NotFound, "order not found")
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
	// Validation
	if req.UserId <= 0 {
		s.logger.Warnw("invalid user_id in ListOrders", "user_id", req.UserId)
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}

	// Apply sensible defaults and limits
	limit := uint64(req.Limit) // #nosec G115 - req.Limit is int32, safe range for uint64
	if limit == 0 {
		limit = 10 // default
	} else if limit > 100 {
		limit = 100 // max limit
	}

	offset := uint64(req.Offset) // #nosec G115 - req.Offset is int32, safe range for uint64

	orders, err := s.svc.ListOrdersByUser(ctx, req.UserId, limit, offset)
	if err != nil {
		s.logger.Errorw("list orders failed", "user_id", req.UserId, "err", err)
		return nil, status.Error(codes.Internal, "failed to list orders")
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
