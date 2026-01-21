package products

import (
	"context"
	"errors"
	"testing"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	proto "github.com/vsespontanno/eCommerce/proto/products"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/apperrors"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/domain/products/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	logger.InitLogger()
}

// MockProducts is a mock implementation of Products interface
type MockProducts struct {
	GetProductByIDFunc  func(ctx context.Context, id int64) (*entity.Product, error)
	GetProductsByIDFunc func(ctx context.Context, ids []int64) ([]*entity.Product, error)
}

func (m *MockProducts) GetProductByID(ctx context.Context, id int64) (*entity.Product, error) {
	return m.GetProductByIDFunc(ctx, id)
}

func (m *MockProducts) GetProductsByID(ctx context.Context, ids []int64) ([]*entity.Product, error) {
	return m.GetProductsByIDFunc(ctx, ids)
}

func TestProductServer_GetProductByID(t *testing.T) {
	tests := []struct {
		name            string
		req             *proto.GetProductByIDRequest
		mockProducts    func() *MockProducts
		expectedProduct *proto.Product
		expectedCode    codes.Code
	}{
		{
			name: "Success",
			req:  &proto.GetProductByIDRequest{Id: 1},
			mockProducts: func() *MockProducts {
				return &MockProducts{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return &entity.Product{ID: 1, Name: "Product 1", Price: 100}, nil
					},
				}
			},
			expectedProduct: &proto.Product{Id: 1, Name: "Product 1", Price: 100},
			expectedCode:    codes.OK,
		},
		{
			name: "Not Found",
			req:  &proto.GetProductByIDRequest{Id: 1},
			mockProducts: func() *MockProducts {
				return &MockProducts{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return nil, apperrors.ErrNoProductFound
					},
				}
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal Error",
			req:  &proto.GetProductByIDRequest{Id: 1},
			mockProducts: func() *MockProducts {
				return &MockProducts{
					GetProductByIDFunc: func(ctx context.Context, id int64) (*entity.Product, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ProductServer{
				products: tt.mockProducts(),
				log:      logger.Log,
			}

			resp, err := server.GetProductByID(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Product.Id != tt.expectedProduct.Id {
					t.Errorf("Expected ID %d, got %d", tt.expectedProduct.Id, resp.Product.Id)
				}
			}
		})
	}
}

func TestProductServer_GetProducts(t *testing.T) {
	tests := []struct {
		name             string
		req              *proto.GetProductsByIDRequest
		mockProducts     func() *MockProducts
		expectedProducts []*proto.Product
		expectedCode     codes.Code
	}{
		{
			name: "Success",
			req:  &proto.GetProductsByIDRequest{Ids: []int64{1, 2}},
			mockProducts: func() *MockProducts {
				return &MockProducts{
					GetProductsByIDFunc: func(ctx context.Context, ids []int64) ([]*entity.Product, error) {
						return []*entity.Product{
							{ID: 1, Name: "Product 1"},
							{ID: 2, Name: "Product 2"},
						}, nil
					},
				}
			},
			expectedProducts: []*proto.Product{
				{Id: 1, Name: "Product 1"},
				{Id: 2, Name: "Product 2"},
			},
			expectedCode: codes.OK,
		},
		{
			name: "Internal Error",
			req:  &proto.GetProductsByIDRequest{Ids: []int64{1}},
			mockProducts: func() *MockProducts {
				return &MockProducts{
					GetProductsByIDFunc: func(ctx context.Context, ids []int64) ([]*entity.Product, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &ProductServer{
				products: tt.mockProducts(),
				log:      logger.Log,
			}

			resp, err := server.GetProducts(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(resp.Products) != len(tt.expectedProducts) {
					t.Errorf("Expected %d products, got %d", len(tt.expectedProducts), len(resp.Products))
				}
			}
		})
	}
}
