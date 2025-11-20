package products

import (
	"context"

	"github.com/vsespontanno/eCommerce/products-service/internal/domain/products/entity"
)

type ProductsStorage interface {
	GetProductByID(ctx context.Context, id int64) (entity.Product, error)
}

type ProductService struct {
	storage ProductsStorage
}

func NewProductService(storage ProductsStorage) *ProductService {
	return &ProductService{
		storage: storage,
	}
}

func (s *ProductService) GetProductByID(ctx context.Context, id int64) (entity.Product, error) {
	return s.storage.GetProductByID(ctx, id)
}

func (s *ProductService) GetProducts(ctx context.Context, ids []int64) ([]entity.Product, error) {
	products := make([]entity.Product, 0, len(ids))
	for _, id := range ids {
		product, err := s.GetProductByID(ctx, id)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}
