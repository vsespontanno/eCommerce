package saga

import (
	"context"

	proto "github.com/vsespontanno/eCommerce/proto/products"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/grpc/dto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Reserver interface {
	Reserve(ctx context.Context, products []*dto.ItemRequest) error
	Release(ctx context.Context, products []*dto.ItemRequest) error
	Commit(ctx context.Context, products []*dto.ItemRequest) error
}

type Server struct {
	reserver Reserver
	proto.UnimplementedSagaProductsServer
}

func NewSagaServer(reserver Reserver) *Server {
	return &Server{
		reserver: reserver,
	}
}

func (s *Server) ReserveProducts(ctx context.Context, req *proto.ReserveProductsRequest) (*proto.ReserveProductsResponse, error) {
	products := mapProtoToDTO(req.Products)
	err := s.reserver.Reserve(ctx, products)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve products: %v", err)

	}
	return &proto.ReserveProductsResponse{Success: true}, nil
}

func (s *Server) ReleaseProducts(ctx context.Context, req *proto.ReleaseProductsRequest) (*proto.ReleaseProductsResponse, error) {
	products := mapProtoToDTO(req.Products)
	err := s.reserver.Release(ctx, products)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to release products: %v", err)

	}
	return &proto.ReleaseProductsResponse{Success: true}, nil
}

func (s *Server) CommitProducts(ctx context.Context, req *proto.CommitProductsRequest) (*proto.CommitProductsResponse, error) {
	products := mapProtoToDTO(req.Products)
	err := s.reserver.Commit(ctx, products)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit products: %v", err)

	}
	return &proto.CommitProductsResponse{Success: true}, nil
}

func mapProtoToDTO(products []*proto.ProductSaga) []*dto.ItemRequest {
	items := make([]*dto.ItemRequest, 0, len(products))
	for _, p := range products {
		items = append(items, &dto.ItemRequest{
			ProductID: int(p.Id),
			Qty:       int(p.Quantity),
		})
	}
	return items
}
