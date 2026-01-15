package saga

import (
	"context"

	proto "github.com/vsespontanno/eCommerce/proto/products"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/grpc/dto"
	"go.uber.org/zap"
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
	logger   *zap.SugaredLogger
	proto.UnimplementedSagaProductsServer
}

func NewSagaServer(reserver Reserver, logger *zap.SugaredLogger) *Server {
	return &Server{
		reserver: reserver,
		logger:   logger,
	}
}

func (s *Server) ReserveProducts(ctx context.Context, req *proto.ReserveProductsRequest) (*proto.ReserveProductsResponse, error) {
	products := mapProtoToDTO(req.Products)
	s.logger.Infow("Reserving products", "count", len(products))

	err := s.reserver.Reserve(ctx, products)
	if err != nil {
		s.logger.Errorw("Failed to reserve products", "error", err, "count", len(products))
		return nil, status.Errorf(codes.Internal, "failed to reserve products: %v", err)
	}

	s.logger.Infow("Products reserved successfully", "count", len(products))
	return &proto.ReserveProductsResponse{Success: true}, nil
}

func (s *Server) ReleaseProducts(ctx context.Context, req *proto.ReleaseProductsRequest) (*proto.ReleaseProductsResponse, error) {
	products := mapProtoToDTO(req.Products)
	s.logger.Infow("Releasing products", "count", len(products))

	err := s.reserver.Release(ctx, products)
	if err != nil {
		s.logger.Errorw("Failed to release products", "error", err, "count", len(products))
		return nil, status.Errorf(codes.Internal, "failed to release products: %v", err)
	}

	s.logger.Infow("Products released successfully", "count", len(products))
	return &proto.ReleaseProductsResponse{Success: true}, nil
}

func (s *Server) CommitProducts(ctx context.Context, req *proto.CommitProductsRequest) (*proto.CommitProductsResponse, error) {
	products := mapProtoToDTO(req.Products)
	s.logger.Infow("Committing products", "count", len(products))

	err := s.reserver.Commit(ctx, products)
	if err != nil {
		s.logger.Errorw("Failed to commit products", "error", err, "count", len(products))
		return nil, status.Errorf(codes.Internal, "failed to commit products: %v", err)
	}

	s.logger.Infow("Products committed successfully", "count", len(products))
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
