package saga

import (
	"context"
	"fmt"

	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/interfaces"
	"go.uber.org/zap"
)

const (
	// MaxTransactionAmount maximum amount for a single transaction (1,000,000 rubles in kopecks)
	MaxTransactionAmount = 100_000_000
)

type SagaWalletService struct {
	walletRepo interfaces.TransactionWallet
	logger     *zap.SugaredLogger
}

func NewSagaWalletService(walletRepo interfaces.TransactionWallet, logger *zap.SugaredLogger) *SagaWalletService {
	return &SagaWalletService{
		walletRepo: walletRepo,
		logger:     logger,
	}
}

// Reserve reserves funds for a transaction (saga step 1)
func (s *SagaWalletService) Reserve(ctx context.Context, userID int64, amount int64) error {
	// Validate input
	if err := s.validateTransactionAmount(userID, amount); err != nil {
		s.logger.Warnw("Invalid reserve request",
			"userID", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	// Reserve funds (repository will check balance)
	err := s.walletRepo.ReserveMoney(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Failed to reserve funds",
			"userID", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	s.logger.Infow("Funds reserved successfully",
		"userID", userID,
		"amount", amount,
	)

	return nil
}

// Release releases previously reserved funds (saga rollback)
func (s *SagaWalletService) Release(ctx context.Context, userID int64, amount int64) error {
	// Validate input
	if err := s.validateTransactionAmount(userID, amount); err != nil {
		s.logger.Warnw("Invalid release request",
			"userID", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	err := s.walletRepo.ReleaseMoney(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Failed to release funds",
			"userID", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	s.logger.Infow("Funds released successfully",
		"userID", userID,
		"amount", amount,
	)

	return nil
}

// Commit commits reserved funds (saga step 2 - final deduction)
func (s *SagaWalletService) Commit(ctx context.Context, userID int64, amount int64) error {
	// Validate input
	if err := s.validateTransactionAmount(userID, amount); err != nil {
		s.logger.Warnw("Invalid commit request",
			"userID", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	err := s.walletRepo.CommitMoney(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Failed to commit funds",
			"userID", userID,
			"amount", amount,
			"error", err,
		)
		return err
	}

	s.logger.Infow("Funds committed successfully",
		"userID", userID,
		"amount", amount,
	)

	return nil
}

// validateTransactionAmount validates transaction parameters
func (s *SagaWalletService) validateTransactionAmount(userID int64, amount int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID: %d", userID)
	}
	if amount <= 0 {
		return fmt.Errorf("amount must be positive: %d", amount)
	}
	if amount > MaxTransactionAmount {
		return fmt.Errorf("amount exceeds maximum transaction limit: %d > %d", amount, MaxTransactionAmount)
	}
	return nil
}
