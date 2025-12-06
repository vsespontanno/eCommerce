package products

import (
	"context"
	"fmt"

	"github.com/vsespontanno/eCommerce/proto/products"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/domain/product/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductsClient struct {
	client products.SagaProductsClient
	logger *zap.SugaredLogger
	port   string
}

func NewProductsClient(port string, logger *zap.SugaredLogger) ProductsClient {
	address := "localhost:" + port

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("Failed to dial gRPC server %s: %v", address, err)
	}

	client := products.NewSagaProductsClient(conn)
	return ProductsClient{
		client: client,
		port:   port,
		logger: logger,
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
		p.logger.Errorw("Error while reserving products", "error", err, "products", len(productIDs))
		return false, err
	}
	if res == nil {
		p.logger.Errorw("Nil response from ReserveProducts", "products", len(productIDs))
		return false, fmt.Errorf("nil response from products service")
	}
	if !res.Success {
		p.logger.Errorw("Failed to reserve products", "error", res.Error, "products", len(productIDs))
		return false, fmt.Errorf("reserve products failed: %s", res.Error)
	}
	p.logger.Infow("Products reserved successfully", "products", len(productIDs))
	return true, nil
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
		p.logger.Errorw("Error while committing products", "error", err, "products", len(productIDs))
		return false, err
	}
	if res == nil {
		p.logger.Errorw("Nil response from CommitProducts", "products", len(productIDs))
		return false, fmt.Errorf("nil response from products service")
	}
	if !res.Success {
		p.logger.Errorw("Failed to commit products", "error", res.Error, "products", len(productIDs))
		return false, fmt.Errorf("commit products failed: %s", res.Error)
	}
	p.logger.Infow("Products committed successfully", "products", len(productIDs))
	return true, nil
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
		p.logger.Errorw("Error while releasing products", "error", err, "products", len(productIDs))
		return false, err
	}
	if res == nil {
		p.logger.Errorw("Nil response from ReleaseProducts", "products", len(productIDs))
		return false, fmt.Errorf("nil response from products service")
	}
	if !res.Success {
		p.logger.Errorw("Failed to release products", "error", res.Error, "products", len(productIDs))
		return false, fmt.Errorf("release products failed: %s", res.Error)
	}
	p.logger.Infow("Products released successfully", "products", len(productIDs))
	return true, nil
}
