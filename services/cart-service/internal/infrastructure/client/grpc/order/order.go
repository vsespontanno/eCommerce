package order

import (
	"context"
	"fmt"
	"log"

	order "github.com/vsespontanno/eCommerce/proto/orders"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/order/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client order.OrderClient
	logger *zap.SugaredLogger
	addr   string
}

func NewOrderClient(addr string, logger *zap.SugaredLogger) *Client {
	// addr уже содержит полный адрес из ConfigMap
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}
	client := order.NewOrderClient(conn)
	logger.Infow("Connected to Order service as a client", "addr", addr)
	return &Client{
		client: client,
		addr:   addr,
		logger: logger,
	}
}

func (o *Client) CreateOrder(ctx context.Context, orderEvent *entity.OrderEvent) (string, error) {
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
			Quantity:  p.Quantity,
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
