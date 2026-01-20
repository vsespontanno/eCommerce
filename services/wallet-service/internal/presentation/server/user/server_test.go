package user

import (
	"context"
	"errors"
	"testing"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/apperrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	logger.InitLogger()
}

// MockWallet is a mock implementation of Wallet interface
type MockWallet struct {
	GetBalanceFunc             func(ctx context.Context) (int64, error)
	GetBalanceWithReservedFunc func(ctx context.Context) (int64, int64, error)
	CreateWalletFunc           func(ctx context.Context) (bool, string, error)
	UpdateBalanceFunc          func(ctx context.Context, amount int64) error
}

func (m *MockWallet) GetBalance(ctx context.Context) (int64, error) {
	return m.GetBalanceFunc(ctx)
}

func (m *MockWallet) GetBalanceWithReserved(ctx context.Context) (int64, int64, error) {
	return m.GetBalanceWithReservedFunc(ctx)
}

func (m *MockWallet) CreateWallet(ctx context.Context) (bool, string, error) {
	return m.CreateWalletFunc(ctx)
}

func (m *MockWallet) UpdateBalance(ctx context.Context, amount int64) error {
	return m.UpdateBalanceFunc(ctx, amount)
}

func TestWalletServer_CreateWallet(t *testing.T) {
	tests := []struct {
		name            string
		mockWallet      func() *MockWallet
		expectedSuccess bool
		expectedMsg     string
		expectedCode    codes.Code
	}{
		{
			name: "Success",
			mockWallet: func() *MockWallet {
				return &MockWallet{
					CreateWalletFunc: func(ctx context.Context) (bool, string, error) {
						return true, "wallet created", nil
					},
				}
			},
			expectedSuccess: true,
			expectedMsg:     "wallet created",
			expectedCode:    codes.OK,
		},
		{
			name: "Internal Error",
			mockWallet: func() *MockWallet {
				return &MockWallet{
					CreateWalletFunc: func(ctx context.Context) (bool, string, error) {
						return false, "", errors.New("db error")
					},
				}
			},
			expectedSuccess: false,
			expectedMsg:     "",
			expectedCode:    codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &WalletServer{
				userWallet: tt.mockWallet(),
				log:        logger.Log,
			}

			resp, err := server.CreateWallet(context.Background(), &proto.CreateWalletRequest{})

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Success != tt.expectedSuccess {
					t.Errorf("Expected success %v, got %v", tt.expectedSuccess, resp.Success)
				}
				if resp.Message != tt.expectedMsg {
					t.Errorf("Expected message %s, got %s", tt.expectedMsg, resp.Message)
				}
			}
		})
	}
}

func TestWalletServer_Balance(t *testing.T) {
	tests := []struct {
		name             string
		mockWallet       func() *MockWallet
		expectedBalance  int64
		expectedReserved int64
		expectedCode     codes.Code
	}{
		{
			name: "Success",
			mockWallet: func() *MockWallet {
				return &MockWallet{
					GetBalanceWithReservedFunc: func(ctx context.Context) (int64, int64, error) {
						return 1000, 100, nil
					},
				}
			},
			expectedBalance:  1000,
			expectedReserved: 100,
			expectedCode:     codes.OK,
		},
		{
			name: "No Wallet",
			mockWallet: func() *MockWallet {
				return &MockWallet{
					GetBalanceWithReservedFunc: func(ctx context.Context) (int64, int64, error) {
						return 0, 0, apperrors.ErrNoWallet
					},
				}
			},
			expectedBalance:  0,
			expectedReserved: 0,
			expectedCode:     codes.OK, // Returns empty balance with message
		},
		{
			name: "Internal Error",
			mockWallet: func() *MockWallet {
				return &MockWallet{
					GetBalanceWithReservedFunc: func(ctx context.Context) (int64, int64, error) {
						return 0, 0, errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &WalletServer{
				userWallet: tt.mockWallet(),
				log:        logger.Log,
			}

			resp, err := server.Balance(context.Background(), &proto.BalanceRequest{})

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Balance != tt.expectedBalance {
					t.Errorf("Expected balance %d, got %d", tt.expectedBalance, resp.Balance)
				}
				if resp.Reserved != tt.expectedReserved {
					t.Errorf("Expected reserved %d, got %d", tt.expectedReserved, resp.Reserved)
				}
			}
		})
	}
}

func TestWalletServer_TopUp(t *testing.T) {
	tests := []struct {
		name            string
		req             *proto.TopUpRequest
		mockWallet      func() *MockWallet
		expectedSuccess bool
		expectedBalance int64
		expectedCode    codes.Code
	}{
		{
			name: "Success",
			req:  &proto.TopUpRequest{Amount: 100},
			mockWallet: func() *MockWallet {
				return &MockWallet{
					UpdateBalanceFunc: func(ctx context.Context, amount int64) error {
						return nil
					},
					GetBalanceWithReservedFunc: func(ctx context.Context) (int64, int64, error) {
						return 1100, 0, nil
					},
				}
			},
			expectedSuccess: true,
			expectedBalance: 1100,
			expectedCode:    codes.OK,
		},
		{
			name:         "Empty Request",
			req:          nil,
			mockWallet:   func() *MockWallet { return &MockWallet{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Zero Amount",
			req:          &proto.TopUpRequest{Amount: 0},
			mockWallet:   func() *MockWallet { return &MockWallet{} },
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "No Wallet",
			req:  &proto.TopUpRequest{Amount: 100},
			mockWallet: func() *MockWallet {
				return &MockWallet{
					UpdateBalanceFunc: func(ctx context.Context, amount int64) error {
						return apperrors.ErrNoWallet
					},
				}
			},
			expectedSuccess: false,
			expectedCode:    codes.OK, // Returns success=false with message
		},
		{
			name: "Validation Error - Negative",
			req:  &proto.TopUpRequest{Amount: -100},
			mockWallet: func() *MockWallet {
				return &MockWallet{
					UpdateBalanceFunc: func(ctx context.Context, amount int64) error {
						return errors.New("amount must be positive")
					},
				}
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Internal Error",
			req:  &proto.TopUpRequest{Amount: 100},
			mockWallet: func() *MockWallet {
				return &MockWallet{
					UpdateBalanceFunc: func(ctx context.Context, amount int64) error {
						return errors.New("db error")
					},
				}
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &WalletServer{
				userWallet: tt.mockWallet(),
				log:        logger.Log,
			}

			resp, err := server.TopUp(context.Background(), tt.req)

			if tt.expectedCode != codes.OK {
				if status.Code(err) != tt.expectedCode {
					t.Errorf("Expected code %v, got %v", tt.expectedCode, status.Code(err))
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp.Success != tt.expectedSuccess {
					t.Errorf("Expected success %v, got %v", tt.expectedSuccess, resp.Success)
				}
				if tt.expectedSuccess && resp.NewBalance != tt.expectedBalance {
					t.Errorf("Expected new balance %d, got %d", tt.expectedBalance, resp.NewBalance)
				}
			}
		})
	}
}
