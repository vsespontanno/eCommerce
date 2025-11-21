package order

import (
	"context"
	"log"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/order/entity"
	proto "github.com/vsespontanno/eCommerce/proto/orders"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	client proto.OrderClient
	logger *zap.SugaredLogger
	port   string
}

func NewOrderClient(port string, logger *zap.SugaredLogger) *OrderClient {
	addr := "localhost:" + port
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}
	client := proto.NewOrderClient(conn)
	logger.Infow("Connected to Products service as a client")
	return &OrderClient{
		client: client,
		port:   port,
		logger: logger,
	}
}

// need to implement
func (o *OrderClient) CreateOrder(ctx context.Context, order *entity.OrderEvent) (string, error) {
	items := make([]*proto.OrderItem, 0, len(order.Products))
	for _, p := range order.Products {
		items = append(items, &proto.OrderItem{
			ProductId: p.ID,
			Quantity:  p.Quantity,
		})
	}
	newOrder := &proto.OrderEvent{
		OrderId: order.OrderID,
		UserId:  order.UserID,
		Status:  order.Status,
		Items:   items,
		Total:   order.Total,
	}

	req := &proto.CreateOrderRequest{
		Order: newOrder,
	}
	resp, err := o.client.CreateOrder(ctx, req)
	if err != nil {
		return resp.Error, err
	}
	return resp.OrderId, nil
}
