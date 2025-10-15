package saga

import (
	"context"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/apperrors"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/interfaces"
	"go.uber.org/zap"
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

func (s *SagaWalletService) Reserve(ctx context.Context, userID int64, amount int64) error {
	userBalance, err := s.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Errorw("Error getting balance", "error", err, "stage", "SagaWalletService.Reserve")
		return err
	}
	if userBalance < amount {
		s.logger.Errorw("Insufficient funds", "stage", "SagaWalletService.Reserve")
		return apperrors.ErrInsufficientFunds
	}
	return s.walletRepo.ReserveMoney(ctx, userID, amount)
}

func (s *SagaWalletService) Release(ctx context.Context, userID int64, amount int64) error {
	err := s.walletRepo.ReleaseMoney(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Error releasing funds", "error", err, "stage", "SagaWalletService.Release")
		return err
	}
	return nil
}

func (s *SagaWalletService) Commit(ctx context.Context, userID int64, amount int64) error {
	err := s.walletRepo.CommitMoney(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Error committing funds", "error", err, "stage", "SagaWalletService.Commit")
		return err
	}
	return nil
}
