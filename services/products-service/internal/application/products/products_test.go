package products

import (
	"context"
	"errors"
	"testing"

	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/products/entity"
)

// MockStorage is a mock implementation of Storage interface
type MockStorage struct {
	GetProductByIDFunc func(ctx context.Context, id int64) (entity.Product, error)
}

func (m *MockStorage) GetProductByID(ctx context.Context, id int64) (entity.Product, error) {
	return m.GetProductByIDFunc(ctx, id)
}

func TestProductService_GetProductByID(t *testing.T) {
	tests := []struct {
		name            string
		id              int64
		mockStorage     func() *MockStorage
		expectedProduct entity.Product
		expectedError   error
	}{
		{
			name: "Success",
			id:   1,
			mockStorage: func() *MockStorage {
				return &MockStorage{
					GetProductByIDFunc: func(ctx context.Context, id int64) (entity.Product, error) {
						return entity.Product{ID: 1, Name: "Product 1"}, nil
					},
				}
			},
			expectedProduct: entity.Product{ID: 1, Name: "Product 1"},
			expectedError:   nil,
		},
		{
			name: "Not Found",
			id:   1,
			mockStorage: func() *MockStorage {
				return &MockStorage{
					GetProductByIDFunc: func(ctx context.Context, id int64) (entity.Product, error) {
						return entity.Product{}, errors.New("not found")
					},
				}
			},
			expectedProduct: entity.Product{},
			expectedError:   errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewProductService(tt.mockStorage())

			product, err := service.GetProductByID(context.Background(), tt.id)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if product != tt.expectedProduct {
					t.Errorf("Expected product %v, got %v", tt.expectedProduct, product)
				}
			}
		})
	}
}

func TestProductService_GetProducts(t *testing.T) {
	tests := []struct {
		name             string
		ids              []int64
		mockStorage      func() *MockStorage
		expectedProducts []entity.Product
		expectedError    error
	}{
		{
			name: "Success",
			ids:  []int64{1, 2},
			mockStorage: func() *MockStorage {
				return &MockStorage{
					GetProductByIDFunc: func(ctx context.Context, id int64) (entity.Product, error) {
						if id == 1 {
							return entity.Product{ID: 1, Name: "Product 1"}, nil
						}
						if id == 2 {
							return entity.Product{ID: 2, Name: "Product 2"}, nil
						}
						return entity.Product{}, errors.New("not found")
					},
				}
			},
			expectedProducts: []entity.Product{
				{ID: 1, Name: "Product 1"},
				{ID: 2, Name: "Product 2"},
			},
			expectedError: nil,
		},
		{
			name: "One Not Found",
			ids:  []int64{1, 3},
			mockStorage: func() *MockStorage {
				return &MockStorage{
					GetProductByIDFunc: func(ctx context.Context, id int64) (entity.Product, error) {
						if id == 1 {
							return entity.Product{ID: 1, Name: "Product 1"}, nil
						}
						return entity.Product{}, errors.New("not found")
					},
				}
			},
			expectedProducts: nil,
			expectedError:    errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewProductService(tt.mockStorage())

			products, err := service.GetProducts(context.Background(), tt.ids)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(products) != len(tt.expectedProducts) {
					t.Errorf("Expected %d products, got %d", len(tt.expectedProducts), len(products))
				}
				for i, p := range products {
					if p != tt.expectedProducts[i] {
						t.Errorf("Expected product %v at index %d, got %v", tt.expectedProducts[i], i, p)
					}
				}
			}
		})
	}
}
