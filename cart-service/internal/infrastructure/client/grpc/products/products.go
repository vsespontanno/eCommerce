package products

import (
	"context"
	"log"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/cart/entity"
	products "github.com/vsespontanno/eCommerce/proto/products"
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
	res, err := c.client.GetProductByID(context.Background(), &products.GetProductByIDRequest{Id: id})
	if err != nil {
		c.logger.Errorf("error while getting product: %w", err.Error())
		return &entity.CartItem{}, err
	}
	var product entity.CartItem
	if res.Product == nil {
		return &entity.CartItem{}, apperrors.ErrProductIsNotInStock
	}
	product.ProductID = res.Product.Id
	product.Price = res.Product.Price

	return &product, nil
}
