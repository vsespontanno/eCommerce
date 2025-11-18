package ordergrpc

import (
	"context"

	"github.com/vsespontanno/eCommerce/order-service/internal/domain/order/entity"
	proto "github.com/vsespontanno/eCommerce/proto/orders"
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

// func (s *Server) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
// 	o, err := s.svc.GetOrder(ctx, req.OrderId)
// 	if err != nil {
// 		s.logger.Errorw("get order failed", "err", err)
// 		return nil, err
// 	}
// 	if o == nil {
// 		return &proto.GetOrderResponse{Error: "not found"}, nil
// 	}

// 	resp := &proto.GetOrderResponse{
// 		Order.Order: o.OrderID,
// 		UserId:  o.UserID,
// 		Total:   o.Total,
// 		Status:  o.Status,
// 	}

// 	for _, it := range o.Products {
// 		resp.Items = append(resp.Items, &proto.OrderItem{
// 			ProductId: it.ProductID,
// 			Quantity:  it.Quantity,
// 		})
// 	}

// 	return resp, nil
// }

// func (s *Server) ListOrders(ctx context.Context, req *proto.ListOrdersRequest) (*proto.ListOrdersResponse, error) {
// 	orders, err := s.svc.ListOrdersByUser(ctx, req.UserId, int(req.Limit), int(req.Offset))
// 	if err != nil {
// 		return nil, err
// 	}
// 	resp := &proto.ListOrdersResponse{}
// 	for _, o := range orders {
// 		r := &proto.GetOrderResponse{
// 			OrderId:   o.ID,
// 			UserId:    o.UserID,
// 			Total:     o.Total,
// 			Status:    o.Status,
// 			CreatedAt: o.CreatedAt.Unix(),
// 		}
// 		for _, it := range o.Items {
// 			r.Items = append(r.Items, &proto.OrderItem{
// 				ProductId: it.ProductID,
// 				Price:     it.Price,
// 				Quantity:  it.Quantity,
// 			})
// 		}
// 		resp.Orders = append(resp.Orders, r)
// 	}
// 	return resp, nil
// }
