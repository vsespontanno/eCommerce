package products

import (
	"context"
	"log"

	"github.com/vsespontanno/eCommerce/cart-service/internal/domain/models"
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

func (c *ProductsClient) Product(ctx context.Context, id int64) (*models.Product, error) {
	res, err := c.client.GetProductByID(context.Background(), &products.GetProductByIDRequest{Id: id})
	if err != nil {
		c.logger.Errorf("error while getting product: %w", err.Error())
		return nil, err
	}
	var product models.Product
	if res.Product == nil {
		return &product, nil
	}
	product.Description = res.Product.Description
	product.ID = res.Product.Id
	product.Name = res.Product.Name
	product.Price = res.Product.Price
	product.Quantity = 1

	return &product, nil
}
