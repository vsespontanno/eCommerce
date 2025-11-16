package user

import (
	"context"

	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/apperrors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserWalletServer struct {
	proto.UnimplementedWalletTopUPServer
	userWallet UserWallet
	log        *zap.SugaredLogger
}

type UserWallet interface {
	GetBalance(ctx context.Context) (int64, error)
	CreateWallet(ctx context.Context) (bool, string, error)
	UpdateBalance(ctx context.Context, amount int64) error
}

func NewUserWalletServer(
	gRPCServer *grpc.Server,
	userWallet UserWallet,
	log *zap.SugaredLogger,
) {
	log.Infow("Registering UserWalletServer")
	proto.RegisterWalletTopUPServer(
		gRPCServer,
		&UserWalletServer{
			userWallet: userWallet,
			log:        log,
		},
	)
}

func (s *UserWalletServer) CreateWallet(ctx context.Context, req *proto.CreateWalletRequest) (*proto.CreateWalletResponse, error) {
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

func (s *UserWalletServer) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	balance, err := s.userWallet.GetBalance(ctx)
	if err != nil {
		if err == apperrors.ErrNoWallet {
			return &proto.BalanceResponse{
				Balance: 0,
				Error:   err.Error(),
			}, nil
		}

		s.log.Errorw("Balance failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to get balance")
	}

	return &proto.BalanceResponse{Balance: balance}, nil
}

func (s *UserWalletServer) TopUp(ctx context.Context, req *proto.TopUpRequest) (*proto.TopUpResponse, error) {
	err := s.userWallet.UpdateBalance(ctx, req.Amount)
	if err != nil {
		if err == apperrors.ErrNoWallet {
			return &proto.TopUpResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		s.log.Errorw("TopUp failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to top up wallet")
	}

	return &proto.TopUpResponse{Success: true}, nil
}
