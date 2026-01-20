package user

import (
	"context"
	"errors"
	"testing"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/apperrors"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/keys"
)

func init() {
	logger.InitLogger()
}

// MockUserWallet is a mock implementation of interfaces.UserWallet
type MockUserWallet struct {
	GetBalanceFunc             func(ctx context.Context, userID int64) (int64, error)
	GetBalanceWithReservedFunc func(ctx context.Context, userID int64) (int64, int64, error)
	UpdateBalanceFunc          func(ctx context.Context, userID int64, amount int64) error
	CreateWalletFunc           func(ctx context.Context, userID int64) (bool, string, error)
}

func (m *MockUserWallet) GetBalance(ctx context.Context, userID int64) (int64, error) {
	return m.GetBalanceFunc(ctx, userID)
}

func (m *MockUserWallet) GetBalanceWithReserved(ctx context.Context, userID int64) (int64, int64, error) {
	return m.GetBalanceWithReservedFunc(ctx, userID)
}

func (m *MockUserWallet) UpdateBalance(ctx context.Context, userID int64, amount int64) error {
	return m.UpdateBalanceFunc(ctx, userID, amount)
}

func (m *MockUserWallet) CreateWallet(ctx context.Context, userID int64) (bool, string, error) {
	return m.CreateWalletFunc(ctx, userID)
}

// MockSSOValidator is a mock implementation of SSOValidator
type MockSSOValidator struct {
	ValidateTokenFunc func(ctx context.Context, token string) (int64, error)
}

func (m *MockSSOValidator) ValidateToken(ctx context.Context, token string) (int64, error) {
	return m.ValidateTokenFunc(ctx, token)
}

func TestWalletService_GetBalance(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		mockSSO       func() *MockSSOValidator
		mockRepo      func() *MockUserWallet
		expectedBal   int64
		expectedError error
	}{
		{
			name:  "Success",
			token: "valid_token",
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{
					ValidateTokenFunc: func(ctx context.Context, token string) (int64, error) {
						return 1, nil
					},
				}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{
					GetBalanceFunc: func(ctx context.Context, userID int64) (int64, error) {
						return 1000, nil
					},
				}
			},
			expectedBal:   1000,
			expectedError: nil,
		},
		{
			name:  "Invalid Token",
			token: "invalid_token",
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{
					ValidateTokenFunc: func(ctx context.Context, token string) (int64, error) {
						return 0, errors.New("invalid token")
					},
				}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{}
			},
			expectedBal:   0,
			expectedError: apperrors.ErrNotAuthorized,
		},
		{
			name:  "Repo Error",
			token: "valid_token",
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{
					ValidateTokenFunc: func(ctx context.Context, token string) (int64, error) {
						return 1, nil
					},
				}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{
					GetBalanceFunc: func(ctx context.Context, userID int64) (int64, error) {
						return 0, errors.New("db error")
					},
				}
			},
			expectedBal:   0,
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewWalletService(tt.mockRepo(), tt.mockSSO(), logger.Log)
			ctx := context.WithValue(context.Background(), keys.JwtKey, tt.token)

			balance, err := service.GetBalance(ctx)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if balance != tt.expectedBal {
					t.Errorf("Expected balance %d, got %d", tt.expectedBal, balance)
				}
			}
		})
	}
}

func TestWalletService_UpdateBalance(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		amount        int64
		mockSSO       func() *MockSSOValidator
		mockRepo      func() *MockUserWallet
		expectedError error
	}{
		{
			name:   "Success",
			token:  "valid_token",
			amount: 100,
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{
					ValidateTokenFunc: func(ctx context.Context, token string) (int64, error) {
						return 1, nil
					},
				}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{
					GetBalanceFunc: func(ctx context.Context, userID int64) (int64, error) {
						return 0, nil // Wallet exists
					},
					UpdateBalanceFunc: func(ctx context.Context, userID int64, amount int64) error {
						return nil
					},
				}
			},
			expectedError: nil,
		},
		{
			name:   "Invalid Amount - Negative",
			token:  "valid_token",
			amount: -100,
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{}
			},
			expectedError: errors.New("amount must be positive, got: -100"),
		},
		{
			name:   "Invalid Amount - Too Large",
			token:  "valid_token",
			amount: MaxTopUpAmount + 1,
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{}
			},
			expectedError: errors.New("amount too large: maximum is 10000000 kopecks (100,000 rubles), got: 10000001"),
		},
		{
			name:   "Wallet Not Found",
			token:  "valid_token",
			amount: 100,
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{
					ValidateTokenFunc: func(ctx context.Context, token string) (int64, error) {
						return 1, nil
					},
				}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{
					GetBalanceFunc: func(ctx context.Context, userID int64) (int64, error) {
						return 0, errors.New("wallet not found")
					},
				}
			},
			expectedError: errors.New("wallet not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewWalletService(tt.mockRepo(), tt.mockSSO(), logger.Log)
			ctx := context.WithValue(context.Background(), keys.JwtKey, tt.token)

			err := service.UpdateBalance(ctx, tt.amount)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestWalletService_CreateWallet(t *testing.T) {
	tests := []struct {
		name            string
		token           string
		mockSSO         func() *MockSSOValidator
		mockRepo        func() *MockUserWallet
		expectedSuccess bool
		expectedMsg     string
		expectedError   error
	}{
		{
			name:  "Success",
			token: "valid_token",
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{
					ValidateTokenFunc: func(ctx context.Context, token string) (int64, error) {
						return 1, nil
					},
				}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{
					CreateWalletFunc: func(ctx context.Context, userID int64) (bool, string, error) {
						return true, "wallet created", nil
					},
				}
			},
			expectedSuccess: true,
			expectedMsg:     "wallet created",
			expectedError:   nil,
		},
		{
			name:  "Repo Error",
			token: "valid_token",
			mockSSO: func() *MockSSOValidator {
				return &MockSSOValidator{
					ValidateTokenFunc: func(ctx context.Context, token string) (int64, error) {
						return 1, nil
					},
				}
			},
			mockRepo: func() *MockUserWallet {
				return &MockUserWallet{
					CreateWalletFunc: func(ctx context.Context, userID int64) (bool, string, error) {
						return false, "", errors.New("db error")
					},
				}
			},
			expectedSuccess: false,
			expectedMsg:     "",
			expectedError:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewWalletService(tt.mockRepo(), tt.mockSSO(), logger.Log)
			ctx := context.WithValue(context.Background(), keys.JwtKey, tt.token)

			success, msg, err := service.CreateWallet(ctx)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if success != tt.expectedSuccess {
					t.Errorf("Expected success %v, got %v", tt.expectedSuccess, success)
				}
				if msg != tt.expectedMsg {
					t.Errorf("Expected message %s, got %s", tt.expectedMsg, msg)
				}
			}
		})
	}
}
