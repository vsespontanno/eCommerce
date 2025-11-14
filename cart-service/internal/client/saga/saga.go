package saga

import (
	"context"
	"log"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
	"github.com/vsespontanno/eCommerce/proto/saga"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SagaClient struct {
	client saga.SagaClient
	logger *zap.SugaredLogger
	port   string
}

func NewSagaClient(port string, logger *zap.SugaredLogger) *SagaClient {
	addr := "localhost:" + port
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}
	client := saga.NewSagaClient(conn)
	logger.Infow("Connected to Saga service as a client")
	return &SagaClient{
		client: client,
		port:   port,
		logger: logger,
	}
}

func (s *SagaClient) StartCheckout(ctx context.Context, userID int64, cart *models.Cart) (string, error) {
	// конвертируем []models.Product → []*saga.Cart
	items := make([]*saga.Cart, 0, len(cart.Items))
	for _, p := range cart.Items {
		items = append(items, &saga.Cart{
			ProductID: p.ProductID,
			Price:     p.Price,
			Quantity:  p.Quantity,
		})
	}

	resp, err := s.client.StartCheckout(ctx, &saga.StartCheckoutRequest{
		UserID: userID,
		Cart:   items,
	})
	if err != nil {
		s.logger.Errorw("Error while starting checkout", "error", err, "stage", "StartCheckout")
		return "", err
	}

	return resp.OrderID, nil
}
