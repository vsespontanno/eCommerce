package products

import (
	"context"
	"log"

	"github.com/vsespontanno/eCommerce/order-service/internal/domain/product/entity"
	"github.com/vsespontanno/eCommerce/proto/products"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductsClient struct {
	client products.SagaProductsClient
	logger *zap.SugaredLogger
	port   string
}

func NewProductsClient(port string) ProductsClient {
	address := "localhost:" + port

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", address, err)
	}

	client := products.NewSagaProductsClient(conn)
	return ProductsClient{
		client: client,
		port:   port,
	}
}

func (p *ProductsClient) ReserveProducts(ctx context.Context, productIDs []entity.Product) (bool, error) {
	req := &products.ReserveProductsRequest{}
	for _, v := range productIDs {
		req.Products = append(req.Products, &products.ProductSaga{
			Id:       int64(v.ID),
			Quantity: int64(v.Quantity),
		})
	}
	res, err := p.client.ReserveProducts(ctx, req)
	if err != nil {
		p.logger.Errorf("error while reserving products: %w", err.Error())
		return false, err
	}
	return res.Success, nil
}

func (p *ProductsClient) CommitProducts(ctx context.Context, productIDs []entity.Product) (bool, error) {
	req := &products.CommitProductsRequest{}
	for _, v := range productIDs {
		req.Products = append(req.Products, &products.ProductSaga{
			Id:       int64(v.ID),
			Quantity: int64(v.Quantity),
		})
	}
	res, err := p.client.CommitProducts(ctx, req)
	if err != nil {
		p.logger.Errorf("error while committing products: %w", err.Error())
		return false, err
	}
	return res.Success, nil
}

func (p *ProductsClient) ReleaseProducts(ctx context.Context, productIDs []entity.Product) (bool, error) {
	req := &products.ReleaseProductsRequest{}
	for _, v := range productIDs {
		req.Products = append(req.Products, &products.ProductSaga{
			Id:       int64(v.ID),
			Quantity: int64(v.Quantity),
		})
	}
	res, err := p.client.ReleaseProducts(ctx, req)
	if err != nil {
		p.logger.Errorf("error while releasing products: %w", err.Error())
		return false, err
	}
	return res.Success, nil
}
