package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	proto "github.com/vsespontanno/eCommerce/proto/products"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/grpc/dto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	logger.InitLogger()
}

// MockReserver is a mock implementation of Reserver
type MockReserver struct {
	ReserveFunc func(ctx context.Context, products []*dto.ItemRequest) error
	ReleaseFunc func(ctx context.Context, products []*dto.ItemRequest) error
	CommitFunc  func(ctx context.Context, products []*dto.ItemRequest) error
}

func (m *MockReserver) Reserve(ctx context.Context, products []*dto.ItemRequest) error {
	return m.ReserveFunc(ctx, products)
}
func (m *MockReserver) Release(ctx context.Context, products []*dto.ItemRequest) error {
	return m.ReleaseFunc(ctx, products)
}
func (m *MockReserver) Commit(ctx context.Context, products []*dto.ItemRequest) error {
	return m.CommitFunc(ctx, products)
}

func TestServer_ReserveProducts(t *testing.T) {
	tests := []struct {
		name         string
		req          *proto.ReserveProductsRequest
		mockReserver func() *MockReserver
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &proto.ReserveProductsRequest{
				Products: []*proto.ProductSaga{{Id: 1, Quantity: 1}},
			},
			mockReserver: func() *MockReserver {
				return &MockReserver{
					ReserveFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return nil
					},
				}
			},
			expectedCode: codes.OK,
		},
		{
			name: "Internal Error",
			req: &proto.ReserveProductsRequest{
				Products: []*proto.ProductSaga{{Id: 1, Quantity: 1}},
			},
			mockReserver: func() *MockReserver {
				return &MockReserver{
					ReserveFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewSagaServer(tt.mockReserver(), logger.Log)

			_, err := server.ReserveProducts(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestServer_ReleaseProducts(t *testing.T) {
	tests := []struct {
		name         string
		req          *proto.ReleaseProductsRequest
		mockReserver func() *MockReserver
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &proto.ReleaseProductsRequest{
				Products: []*proto.ProductSaga{{Id: 1, Quantity: 1}},
			},
			mockReserver: func() *MockReserver {
				return &MockReserver{
					ReleaseFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return nil
					},
				}
			},
			expectedCode: codes.OK,
		},
		{
			name: "Internal Error",
			req: &proto.ReleaseProductsRequest{
				Products: []*proto.ProductSaga{{Id: 1, Quantity: 1}},
			},
			mockReserver: func() *MockReserver {
				return &MockReserver{
					ReleaseFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewSagaServer(tt.mockReserver(), logger.Log)

			_, err := server.ReleaseProducts(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestServer_CommitProducts(t *testing.T) {
	tests := []struct {
		name         string
		req          *proto.CommitProductsRequest
		mockReserver func() *MockReserver
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &proto.CommitProductsRequest{
				Products: []*proto.ProductSaga{{Id: 1, Quantity: 1}},
			},
			mockReserver: func() *MockReserver {
				return &MockReserver{
					CommitFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return nil
					},
				}
			},
			expectedCode: codes.OK,
		},
		{
			name: "Internal Error",
			req: &proto.CommitProductsRequest{
				Products: []*proto.ProductSaga{{Id: 1, Quantity: 1}},
			},
			mockReserver: func() *MockReserver {
				return &MockReserver{
					CommitFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewSagaServer(tt.mockReserver(), logger.Log)

			_, err := server.CommitProducts(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
