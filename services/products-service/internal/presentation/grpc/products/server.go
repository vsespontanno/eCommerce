package products

import (
	"context"
	"errors"

	proto "github.com/vsespontanno/eCommerce/proto/products"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/products/entity"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductServer struct {
	proto.UnimplementedProductsServer
	products Products
	log      *zap.SugaredLogger
}

type Products interface {
	GetProductByID(ctx context.Context, id int64) (*entity.Product, error)
	GetProductsByID(ctx context.Context, ids []int64) ([]*entity.Product, error)
}

func NewProductServer(gRPCServer *grpc.Server, products Products, log *zap.SugaredLogger) {
	proto.RegisterProductsServer(gRPCServer, &ProductServer{
		products: products,
		log:      log,
	})
}

func (s *ProductServer) GetProductByID(ctx context.Context, req *proto.GetProductByIDRequest) (*proto.GetProductByIDResponse, error) {
	product, err := s.products.GetProductByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNoProductFound) {
			return nil, status.Errorf(codes.NotFound, "product with ID %d not found", req.Id)
		}

		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}
	return &proto.GetProductByIDResponse{Product: &proto.Product{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}}, nil
}

func (s *ProductServer) GetProducts(ctx context.Context, req *proto.GetProductsByIDRequest) (*proto.GetProductsByIDResponse, error) {
	s.log.Infow("GetProducts request received", "ids", req.Ids, "count", len(req.Ids))

	products, err := s.products.GetProductsByID(ctx, req.Ids)
	if err != nil {
		s.log.Errorw("failed to get products", "error", err, "ids", req.Ids)
		return nil, status.Errorf(codes.Internal, "failed to get products: %v", err)
	}

	var protoProducts []*proto.Product
	for _, product := range products {
		protoProducts = append(protoProducts, &proto.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		})
	}

	s.log.Infow("GetProducts completed", "requested", len(req.Ids), "found", len(protoProducts))
	return &proto.GetProductsByIDResponse{Products: protoProducts}, nil
}
