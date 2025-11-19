package order

import (
	"context"
	"log"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/order/entity"
	order "github.com/vsespontanno/eCommerce/proto/orders"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	client order.OrderClient
	logger *zap.SugaredLogger
	port   string
}

func NewProductsClient(port string, logger *zap.SugaredLogger) *OrderClient {
	addr := "localhost:" + port
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}
	client := order.NewOrderClient(conn)
	logger.Infow("Connected to Products service as a client")
	return &OrderClient{
		client: client,
		port:   port,
		logger: logger,
	}
}

// need to implement
func (o *OrderClient) CreateOrder(ctx context.Context, order *entity.OrderEvent) (string, error) {
	return "no", nil
}
