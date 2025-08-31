package products

import (
	"context"
	"errors"
	"log"

	"github.com/vsespontanno/eCommerce/products-service/internal/domain/models"
	proto "github.com/vsespontanno/eCommerce/proto/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductServer struct {
	proto.UnimplementedProductsServer
	products Products
}

type Products interface {
	GetProductByID(ctx context.Context, id int64) (*models.Product, error)
	GetProductsByID(ctx context.Context, ids []int64) ([]*models.Product, error)
}

func NewProductServer(gRPCServer *grpc.Server, products Products) {
	log.Println("Registering ProductServer")
	proto.RegisterProductsServer(gRPCServer, &ProductServer{products: products})
}

func (s *ProductServer) GetProductByID(ctx context.Context, req *proto.GetProductByIDRequest) (*proto.GetProductByIDResponse, error) {
	log.Printf("GetProduct request received: %v", req)
	product, err := s.products.GetProductByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, models.ErrNoProductFound) {
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
	log.Printf("GetProducts request received: %v", req)
	products, err := s.products.GetProductsByID(ctx, req.Ids)
	if err != nil {
		return nil, err
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

	return &proto.GetProductsByIDResponse{Products: protoProducts}, nil
}
