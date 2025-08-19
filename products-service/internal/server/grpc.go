package server

import (
	"context"
	"log"

	repository "github.com/vsespontanno/eCommerce/products-service/internal/repository/pg"
	"github.com/vsespontanno/eCommerce/proto/products"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	products.UnimplementedProductsServer
	repo *repository.ProductStore
}

func NewGrpcServer(repo *repository.ProductStore) *GrpcServer {
	return &GrpcServer{repo: repo}
}

func (s *GrpcServer) GetProducts(ctx context.Context, req *products.None) (*products.ProductsResponse, error) {
	prods, err := s.repo.GetProducts(ctx)
	if err != nil {
		log.Printf("Error getting products: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve products")
	}

	var pbProducts []*products.Product
	for _, p := range prods {
		pbProducts = append(pbProducts, &products.Product{
			Id:           p.ID,
			Name:         p.Name,
			Price:        p.Price,
			Description:  p.Description,
			Category:     p.Category,
			Brand:        p.Brand,
			Rating:       int32(p.Rating),
			NumReviews:   int32(p.NumReviews),
			CountInStock: int32(p.CountInStock),
		})
	}

	return &products.ProductsResponse{Products: pbProducts}, nil
}
