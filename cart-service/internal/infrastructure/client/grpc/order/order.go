package order

import (
	"context"
	"fmt"
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
	client := order.NewOrderClient(conn)
	logger.Infow("Connected to Order service as a client")
	return &OrderClient{
		client: client,
		port:   port,
		logger: logger,
	}
}

func (o *OrderClient) CreateOrder(ctx context.Context, orderEvent *entity.OrderEvent) (string, error) {
	// Конвертируем entity.OrderEvent в proto.OrderEvent
	protoOrder := &order.OrderEvent{
		OrderId: orderEvent.OrderID,
		UserId:  orderEvent.UserID,
		Total:   orderEvent.Total,
		Status:  orderEvent.Status,
	}

	// Конвертируем Products
	for _, p := range orderEvent.Products {
		protoOrder.Items = append(protoOrder.Items, &order.OrderItem{
			ProductId: p.ID,
			Quantity:  int64(p.Quantity),
		})
	}

	// Вызываем gRPC метод
	resp, err := o.client.CreateOrder(ctx, &order.CreateOrderRequest{
		Order: protoOrder,
	})

	if err != nil {
		o.logger.Errorw("Failed to create order via gRPC", "error", err, "orderID", orderEvent.OrderID)
		return "", err
	}

	if resp.Error != "" {
		o.logger.Errorw("Order service returned error", "error", resp.Error, "orderID", orderEvent.OrderID)
		return "", fmt.Errorf("order service error: %s", resp.Error)
	}

	o.logger.Infow("Order created successfully", "orderID", resp.OrderId)
	return resp.OrderId, nil
}
