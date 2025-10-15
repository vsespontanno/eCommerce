package saga

import (
	"context"

	"github.com/vsespontanno/eCommerce/products-service/internal/grpc/dto"
	proto "github.com/vsespontanno/eCommerce/proto/product"
)

type Reserver interface {
	Reserve(ctx context.Context, products []*dto.ItemRequest) error

	Release(ctx context.Context, userID int64) error
}

type SagaServer struct {
	reserver Reserver
	proto.UnimplementedSagaProductsServer
}

func (s *SagaServer) ReserveProducts(ctx context.Context, req *proto.ReserveProductsRequest) (*proto.ReserveProductsResponse, error) {
	var products []*dto.ItemRequest
	for _, product := range req.Products {
		productRequest := &dto.ItemRequest{
			ProductID: int(product.Id),
			Qty:       int(product.Quantity),
		}
		products = append(products, productRequest)
	}
	err := s.reserver.Reserve(ctx, products)
	if err != nil {
		return nil, err
	}
	return &proto.ReserveProductsResponse{}, nil
}

func (s *SagaServer) ReleaseProducts(ctx context.Context, req *proto.ReleaseProductsRequest) (*proto.ReleaseProductsResponse, error) {
	return &proto.ReleaseProductsResponse{}, nil
}

func (s *SagaServer) CommitProducts(ctx context.Context, req *proto.CommitProductsRequest) (*proto.CommitProductsResponse, error) {
	return &proto.CommitProductsResponse{}, nil
}
