package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/apperrors"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/keys"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/interfaces"
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

func (s *UserWalletService) GetBalance(ctx context.Context) (int64, error) {
	token := ctx.Value(keys.JwtKey).(string)
	userID, err := s.ssoClient.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Errorw("Error validating token", "error", err)
		return 0, apperrors.ErrNotAuthorized
	}
	balance, err := s.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Errorw("Error getting balance", "error", err)
		return 0, err
	}
	return balance, nil
}

func (s *UserWalletService) UpdateBalance(ctx context.Context, amount int64) error {
	token := ctx.Value(keys.JwtKey).(string)
	userID, err := s.ssoClient.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Errorw("Error validating token", "error", err)
		return apperrors.ErrNotAuthorized
	}
	_, err = s.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		s.logger.Errorw("Error getting balance", "error", err)
		if err == apperrors.ErrNoWallet {
			return apperrors.ErrNoWallet
		}
		return err
	}
	err = s.walletRepo.UpdateBalance(ctx, userID, amount)
	if err != nil {
		s.logger.Errorw("Error updating balance", "error", err)
	}
	return err
}

func (s *UserWalletService) CreateWallet(ctx context.Context) (bool, string, error) {
	token := ctx.Value(keys.JwtKey).(string)
	userID, err := s.ssoClient.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Errorw("Error validating token", "error", err)
		return false, "", apperrors.ErrNotAuthorized
	}
	success, message, err := s.walletRepo.CreateWallet(ctx, userID)
	if err != nil {
		s.logger.Errorw("Error creating wallet", "error", err)
		return false, "", err
	}
	return success, message, nil
}
