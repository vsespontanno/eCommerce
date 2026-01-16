package user

import (
	"context"
	"strings"

	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/apperrors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WalletServer struct {
	proto.UnimplementedWalletTopUPServer
	userWallet Wallet
	log        *zap.SugaredLogger
}

type Wallet interface {
	GetBalance(ctx context.Context) (int64, error)
	GetBalanceWithReserved(ctx context.Context) (balance int64, reserved int64, err error)
	CreateWallet(ctx context.Context) (bool, string, error)
	UpdateBalance(ctx context.Context, amount int64) error
}

func NewUserWalletServer(
	gRPCServer *grpc.Server,
	userWallet Wallet,
	log *zap.SugaredLogger,
) {
	log.Infow("Registering WalletServer")
	proto.RegisterWalletTopUPServer(
		gRPCServer,
		&WalletServer{
			userWallet: userWallet,
			log:        log,
		},
	)
}

func (s *WalletServer) CreateWallet(ctx context.Context, req *proto.CreateWalletRequest) (*proto.CreateWalletResponse, error) {
	success, message, err := s.userWallet.CreateWallet(ctx)
	if err != nil {
		s.log.Errorw("CreateWallet failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to create wallet")
	}

	return &proto.CreateWalletResponse{
		Success: success,
		Message: message,
	}, nil
}

func (s *WalletServer) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	balance, reserved, err := s.userWallet.GetBalanceWithReserved(ctx)
	if err != nil {
		if err == apperrors.ErrNoWallet {
			return &proto.BalanceResponse{
				Balance:  0,
				Reserved: 0,
				Message:  err.Error(),
			}, nil
		}

		s.log.Errorw("Balance failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to get balance")
	}

	return &proto.BalanceResponse{
		Balance:  balance,
		Reserved: reserved,
	}, nil
}

func (s *WalletServer) TopUp(ctx context.Context, req *proto.TopUpRequest) (*proto.TopUpResponse, error) {
	if req == nil || req.Amount == 0 {
		s.log.Warnw("TopUp called with empty or zero amount")
		return nil, status.Error(codes.InvalidArgument, "amount is required and must be non-zero")
	}

	err := s.userWallet.UpdateBalance(ctx, req.Amount)
	if err != nil {
		if err == apperrors.ErrNoWallet {
			return &proto.TopUpResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}

		errMsg := err.Error()
		if strings.Contains(errMsg, "amount") && (strings.Contains(errMsg, "too small") || strings.Contains(errMsg, "too large") || strings.Contains(errMsg, "must be positive")) {
			s.log.Warnw("TopUp validation failed", "error", err)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		s.log.Errorw("TopUp failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to top up wallet")
	}

	// Get new balance after top up
	newBalance, _, err := s.userWallet.GetBalanceWithReserved(ctx)
	if err != nil {
		s.log.Errorw("Failed to get balance after top up", "error", err)
		// Still return success since top up succeeded
		return &proto.TopUpResponse{
			Success:    true,
			Message:    "top up successful, but failed to retrieve new balance",
			NewBalance: 0,
		}, nil
	}

	return &proto.TopUpResponse{
		Success:    true,
		Message:    "top up successful",
		NewBalance: newBalance,
	}, nil
}
