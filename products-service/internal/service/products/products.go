package products

import (
	"context"

	"github.com/vsespontanno/eCommerce/products-service/internal/domain/models"
)

type ProductsStorage interface {
	GetProductByID(ctx context.Context, id int64) (models.Product, error)
}

type ProductService struct {
	storage ProductsStorage
}

func NewProductService(storage ProductsStorage) *ProductService {
	return &ProductService{
		storage: storage,
	}
}

func (s *ProductService) GetProductByID(ctx context.Context, id int64) (models.Product, error) {
	return s.storage.GetProductByID(ctx, id)
}

func (s *ProductService) GetProducts(ctx context.Context, ids []int64) ([]models.Product, error) {
	products := make([]models.Product, 0, len(ids))
	for _, id := range ids {
		product, err := s.GetProductByID(ctx, id)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}
