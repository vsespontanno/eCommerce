package ordergrpc

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/vsespontanno/eCommerce/pkg/logger"
	proto "github.com/vsespontanno/eCommerce/proto/orders"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/domain/order/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	logger.InitLogger()
}

// MockOrderSvc is a mock implementation of OrderSvc
type MockOrderSvc struct {
	CreateOrderFunc      func(ctx context.Context, order *entity.Order) (string, error)
	GetOrderFunc         func(ctx context.Context, orderID string) (*entity.Order, error)
	ListOrdersByUserFunc func(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error)
}

func (m *MockOrderSvc) CreateOrder(ctx context.Context, order *entity.Order) (string, error) {
	return m.CreateOrderFunc(ctx, order)
}

func (m *MockOrderSvc) GetOrder(ctx context.Context, orderID string) (*entity.Order, error) {
	return m.GetOrderFunc(ctx, orderID)
}

func (m *MockOrderSvc) ListOrdersByUser(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
	return m.ListOrdersByUserFunc(ctx, userID, limit, offset)
}

func TestServer_CreateOrder(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name         string
		req          *proto.CreateOrderRequest
		mockSvc      func() *MockOrderSvc
		expectedID   string
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req: &proto.CreateOrderRequest{
				Order: &proto.OrderEvent{
					OrderId: validUUID,
					UserId:  1,
					Total:   1000,
					Status:  "pending",
					Items: []*proto.OrderItem{
						{ProductId: 1, Quantity: 2},
					},
				},
			},
			mockSvc: func() *MockOrderSvc {
				return &MockOrderSvc{
					CreateOrderFunc: func(ctx context.Context, order *entity.Order) (string, error) {
						return validUUID, nil
					},
				}
			},
			expectedID:   validUUID,
			expectedCode: codes.OK,
		},
		{
			name:         "Empty Order",
			req:          &proto.CreateOrderRequest{Order: nil},
			mockSvc:      func() *MockOrderSvc { return &MockOrderSvc{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Empty OrderID",
			req: &proto.CreateOrderRequest{
				Order: &proto.OrderEvent{
					OrderId: "",
					UserId:  1,
				},
			},
			mockSvc:      func() *MockOrderSvc { return &MockOrderSvc{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid UUID",
			req: &proto.CreateOrderRequest{
				Order: &proto.OrderEvent{
					OrderId: "invalid-uuid",
					UserId:  1,
				},
			},
			mockSvc:      func() *MockOrderSvc { return &MockOrderSvc{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Invalid UserID",
			req: &proto.CreateOrderRequest{
				Order: &proto.OrderEvent{
					OrderId: validUUID,
					UserId:  0,
				},
			},
			mockSvc:      func() *MockOrderSvc { return &MockOrderSvc{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal Error",
			req: &proto.CreateOrderRequest{
				Order: &proto.OrderEvent{
					OrderId: validUUID,
					UserId:  1,
				},
			},
			mockSvc: func() *MockOrderSvc {
				return &MockOrderSvc{
					CreateOrderFunc: func(ctx context.Context, order *entity.Order) (string, error) {
						return "", errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewGRPCServer(tt.mockSvc(), logger.Log)

			resp, err := server.CreateOrder(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.OrderId != tt.expectedID {
					t.Errorf("Expected ID %s, got %s", tt.expectedID, resp.OrderId)
				}
			}
		})
	}
}

func TestServer_GetOrder(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name          string
		req           *proto.GetOrderRequest
		mockSvc       func() *MockOrderSvc
		expectedOrder *proto.OrderEvent
		expectedCode  codes.Code
	}{
		{
			name: "Success",
			req:  &proto.GetOrderRequest{OrderId: validUUID},
			mockSvc: func() *MockOrderSvc {
				return &MockOrderSvc{
					GetOrderFunc: func(ctx context.Context, orderID string) (*entity.Order, error) {
						return &entity.Order{
							OrderID: validUUID,
							UserID:  1,
							Total:   1000,
							Status:  "completed",
							Products: []entity.OrderItem{
								{ProductID: 1, Quantity: 2},
							},
						}, nil
					},
				}
			},
			expectedOrder: &proto.OrderEvent{
				OrderId: validUUID,
				UserId:  1,
				Total:   1000,
				Status:  "completed",
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Empty OrderID",
			req:          &proto.GetOrderRequest{OrderId: ""},
			mockSvc:      func() *MockOrderSvc { return &MockOrderSvc{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Invalid UUID",
			req:          &proto.GetOrderRequest{OrderId: "invalid-uuid"},
			mockSvc:      func() *MockOrderSvc { return &MockOrderSvc{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Not Found",
			req:  &proto.GetOrderRequest{OrderId: validUUID},
			mockSvc: func() *MockOrderSvc {
				return &MockOrderSvc{
					GetOrderFunc: func(ctx context.Context, orderID string) (*entity.Order, error) {
						return nil, nil // Service returns nil if not found
					},
				}
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Internal Error",
			req:  &proto.GetOrderRequest{OrderId: validUUID},
			mockSvc: func() *MockOrderSvc {
				return &MockOrderSvc{
					GetOrderFunc: func(ctx context.Context, orderID string) (*entity.Order, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewGRPCServer(tt.mockSvc(), logger.Log)

			resp, err := server.GetOrder(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Order.OrderId != tt.expectedOrder.OrderId {
					t.Errorf("Expected OrderID %s, got %s", tt.expectedOrder.OrderId, resp.Order.OrderId)
				}
			}
		})
	}
}

func TestServer_ListOrders(t *testing.T) {
	tests := []struct {
		name         string
		req          *proto.ListOrdersRequest
		mockSvc      func() *MockOrderSvc
		expectedLen  int
		expectedCode codes.Code
	}{
		{
			name: "Success",
			req:  &proto.ListOrdersRequest{UserId: 1, Limit: 10, Offset: 0},
			mockSvc: func() *MockOrderSvc {
				return &MockOrderSvc{
					ListOrdersByUserFunc: func(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
						return []entity.Order{
							{OrderID: "order-1"},
							{OrderID: "order-2"},
						}, nil
					},
				}
			},
			expectedLen:  2,
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid UserID",
			req:          &proto.ListOrdersRequest{UserId: 0},
			mockSvc:      func() *MockOrderSvc { return &MockOrderSvc{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal Error",
			req:  &proto.ListOrdersRequest{UserId: 1},
			mockSvc: func() *MockOrderSvc {
				return &MockOrderSvc{
					ListOrdersByUserFunc: func(ctx context.Context, userID int64, limit, offset uint64) ([]entity.Order, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewGRPCServer(tt.mockSvc(), logger.Log)

			resp, err := server.ListOrders(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(resp.Orders) != tt.expectedLen {
					t.Errorf("Expected %d orders, got %d", tt.expectedLen, len(resp.Orders))
				}
			}
		})
	}
}
