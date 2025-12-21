package products

import (
	"context"
	"log"

	products "github.com/vsespontanno/eCommerce/proto/products"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/domain/cart/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductsClient struct {
	client products.ProductsClient
	logger *zap.SugaredLogger
	port   string
}

func NewProductsClient(port string, logger *zap.SugaredLogger) *ProductsClient {
	addr := "localhost:" + port
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}
	client := products.NewProductsClient(conn)
	logger.Infow("Connected to Products service as a client")
	return &ProductsClient{
		client: client,
		port:   port,
		logger: logger,
	}
}

func (c *ProductsClient) Product(ctx context.Context, id int64) (*entity.CartItem, error) {
	res, err := c.client.GetProductByID(ctx, &products.GetProductByIDRequest{Id: id})
	if err != nil {
		c.logger.Errorf("error while getting product: %v", err)
		return nil, err
	}
	if res.Product == nil {
		return nil, apperrors.ErrProductIsNotInStock
	}

	product := &entity.CartItem{
		ProductID: res.Product.Id,
		Price:     res.Product.Price,
		Quantity:  1,
	}

	return product, nil
}
