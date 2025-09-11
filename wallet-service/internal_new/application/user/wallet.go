package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/vsespontanno/eCommerce/wallet-service/internal_new/domain/wallet/interfaces"
)

type SSOValidator interface {
	ValidateToken(ctx context.Context, token string) (int64, error)
}

type UserWalletService struct {
	walletRepo interfaces.UserWallet
	ssoClient  SSOValidator
	logger     *zap.SugaredLogger
}

func NewWalletService(walletRepo interfaces.UserWallet, ssoClient SSOValidator, logger *zap.SugaredLogger) *UserWalletService {
	return &UserWalletService{
		walletRepo: walletRepo,
		ssoClient:  ssoClient,
		logger:     logger,
	}
}

func (s *UserWalletService) GetBalance(ctx context.Context) (float64, error) {
	token := ctx.Value("jwt_token").(string)
	userID, err := s.ssoClient.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Errorw("Error validating token", "error", err)
		return 0, err
	}
	balance, err := s.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Errorw("Error getting balance", "error", err)
		return 0, err
	}
	return balance, nil
}

func (s *UserWalletService) UpdateBalance(ctx context.Context, amount float64) error {
	token := ctx.Value("jwt_token").(string)
	userID, err := s.ssoClient.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Errorw("Error validating token", "error", err)
		return err
	}
	err = s.walletRepo.UpdateBalance(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Error updating balance", "error", err)
	}
	return err
}
