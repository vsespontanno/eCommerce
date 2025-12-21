package user

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/apperrors"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/keys"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/interfaces"
)

const (
	// MaxTopUpAmount maximum amount for a single top-up (100,000 rubles in kopecks)
	MaxTopUpAmount = 10_000_000
	// MinTopUpAmount minimum amount for a single top-up (1 ruble in kopecks)
	MinTopUpAmount = 100
)

type SSOValidator interface {
	ValidateToken(ctx context.Context, token string) (int64, error)
}

type WalletService struct {
	walletRepo interfaces.UserWallet
	ssoClient  SSOValidator
	logger     *zap.SugaredLogger
}

func NewWalletService(walletRepo interfaces.UserWallet, ssoClient SSOValidator, logger *zap.SugaredLogger) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
		ssoClient:  ssoClient,
		logger:     logger,
	}
}

// GetBalance returns the current balance for the authenticated user
func (s *WalletService) GetBalance(ctx context.Context) (int64, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	balance, err := s.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Errorw("Failed to get balance",
			"userID", userID,
			"error", err,
		)
		return 0, err
	}

	s.logger.Infow("Balance retrieved successfully",
		"userID", userID,
		"balance", balance,
	)

	return balance, nil
}

// UpdateBalance adds funds to the user's wallet (top up)
func (s *WalletService) UpdateBalance(ctx context.Context, amount int64) error {
	// Validate amount
	if err := s.validateTopUpAmount(amount); err != nil {
		s.logger.Warnw("Invalid top-up amount",
			"amount", amount,
			"error", err,
		)
		return err
	}

	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Check if wallet exists
	_, err = s.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Errorw("Failed to verify wallet existence",
			"userID", userID,
			"error", err,
		)
		return err
	}

	// Update balance
	err = s.walletRepo.UpdateBalance(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Failed to update balance",
			"userID", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	s.logger.Infow("Balance updated successfully",
		"userID", userID,
		"amount", amount,
	)

	return nil
}

// CreateWallet creates a new wallet for the authenticated user
func (s *WalletService) CreateWallet(ctx context.Context) (bool, string, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return false, "", err
	}

	success, message, err := s.walletRepo.CreateWallet(ctx, userID)
	if err != nil {
		s.logger.Errorw("Failed to create wallet",
			"userID", userID,
			"error", err,
		)
		return false, "", err
	}

	s.logger.Infow("Wallet created successfully",
		"userID", userID,
	)

	return success, message, nil
}

// getUserIDFromContext extracts and validates user ID from JWT token in context
func (s *WalletService) getUserIDFromContext(ctx context.Context) (int64, error) {
	token, ok := ctx.Value(keys.JwtKey).(string)
	if !ok || token == "" {
		s.logger.Warn("Missing or invalid JWT token in context")
		return 0, apperrors.ErrNotAuthorized
	}

	userID, err := s.ssoClient.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Errorw("Failed to validate token",
			"error", err,
		)
		return 0, apperrors.ErrNotAuthorized
	}

	if userID <= 0 {
		s.logger.Warnw("Invalid user ID from token",
			"userID", userID,
		)
		return 0, apperrors.ErrNotAuthorized
	}

	return userID, nil
}

// validateTopUpAmount validates the top-up amount
func (s *WalletService) validateTopUpAmount(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive, got: %d", amount)
	}
	if amount < MinTopUpAmount {
		return fmt.Errorf("amount too small: minimum is %d kopecks (1 ruble), got: %d", MinTopUpAmount, amount)
	}
	if amount > MaxTopUpAmount {
		return fmt.Errorf("amount too large: maximum is %d kopecks (100,000 rubles), got: %d", MaxTopUpAmount, amount)
	}
	return nil
}
