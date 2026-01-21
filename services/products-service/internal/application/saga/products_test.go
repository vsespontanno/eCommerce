package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/products-service/internal/presentation/grpc/dto"
)

func init() {
	logger.InitLogger()
}

// MockProductStorage is a mock implementation of ProductStorage
type MockProductStorage struct {
	ReserveTxnFunc func(ctx context.Context, products []*dto.ItemRequest) error
	ReleaseTxnFunc func(ctx context.Context, products []*dto.ItemRequest) error
	CommitTxnFunc  func(ctx context.Context, products []*dto.ItemRequest) error
}

func (m *MockProductStorage) ReserveTxn(ctx context.Context, products []*dto.ItemRequest) error {
	return m.ReserveTxnFunc(ctx, products)
}
func (m *MockProductStorage) ReleaseTxn(ctx context.Context, products []*dto.ItemRequest) error {
	return m.ReleaseTxnFunc(ctx, products)
}
func (m *MockProductStorage) CommitTxn(ctx context.Context, products []*dto.ItemRequest) error {
	return m.CommitTxnFunc(ctx, products)
}

func TestService_Reserve(t *testing.T) {
	tests := []struct {
		name        string
		mockStorage func() *MockProductStorage
		expectedErr error
	}{
		{
			name: "Success",
			mockStorage: func() *MockProductStorage {
				return &MockProductStorage{
					ReserveTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name: "Transient Error - Retry Success",
			mockStorage: func() *MockProductStorage {
				attempts := 0
				return &MockProductStorage{
					ReserveTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						attempts++
						if attempts < 3 {
							return errors.New("deadlock detected")
						}
						return nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name: "Transient Error - Max Attempts Reached",
			mockStorage: func() *MockProductStorage {
				return &MockProductStorage{
					ReserveTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return errors.New("deadlock detected")
					},
				}
			},
			// We expect an error, but the exact message depends on implementation details
			// so we just check if error is not nil
			expectedErr: errors.New("reserve failed after 5 attempts"),
		},
		{
			name: "Non-Transient Error",
			mockStorage: func() *MockProductStorage {
				return &MockProductStorage{
					ReserveTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return errors.New("not enough stock")
					},
				}
			},
			expectedErr: errors.New("reserve failed: not enough stock"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSagaService(tt.mockStorage(), logger.Log)

			err := service.Reserve(context.Background(), nil)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_Release(t *testing.T) {
	tests := []struct {
		name        string
		mockStorage func() *MockProductStorage
		expectedErr error
	}{
		{
			name: "Success",
			mockStorage: func() *MockProductStorage {
				return &MockProductStorage{
					ReleaseTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name: "Error",
			mockStorage: func() *MockProductStorage {
				return &MockProductStorage{
					ReleaseTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return errors.New("db error")
					},
				}
			},
			expectedErr: errors.New("release failed: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSagaService(tt.mockStorage(), logger.Log)

			err := service.Release(context.Background(), nil)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_Commit(t *testing.T) {
	tests := []struct {
		name        string
		mockStorage func() *MockProductStorage
		expectedErr error
	}{
		{
			name: "Success",
			mockStorage: func() *MockProductStorage {
				return &MockProductStorage{
					CommitTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return nil
					},
				}
			},
			expectedErr: nil,
		},
		{
			name: "Error",
			mockStorage: func() *MockProductStorage {
				return &MockProductStorage{
					CommitTxnFunc: func(ctx context.Context, products []*dto.ItemRequest) error {
						return errors.New("db error")
					},
				}
			},
			expectedErr: errors.New("commit failed: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSagaService(tt.mockStorage(), logger.Log)

			err := service.Commit(context.Background(), nil)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
