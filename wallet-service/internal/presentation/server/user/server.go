package user

import (
	"context"
	"log"

	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/domain/wallet/entity/apperrors"
	"google.golang.org/grpc"
)

type UserWalletServer struct {
	proto.UnimplementedWalletTopUPServer
	userWallet UserWallet
}

type UserWallet interface {
	GetBalance(ctx context.Context) (int64, error)
	CreateWallet(ctx context.Context) (bool, string, error)
	UpdateBalance(ctx context.Context, amount int64) error
}

func NewUserWalletServer(gRPCServer *grpc.Server, userWallet UserWallet) {
	log.Println("Registering UserWalletServer")
	proto.RegisterWalletTopUPServer(gRPCServer, &UserWalletServer{userWallet: userWallet})
}

func (s *UserWalletServer) CreateWallet(ctx context.Context, req *proto.CreateWalletRequest) (*proto.CreateWalletResponse, error) {
	success, message, err := s.userWallet.CreateWallet(ctx)
	if err != nil {
		return nil, err
	}
	return &proto.CreateWalletResponse{Success: success, Message: message}, nil
}

func (s *UserWalletServer) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	balance, err := s.userWallet.GetBalance(ctx)
	if err != nil {
		if err == apperrors.ErrNoWallet {
			return &proto.BalanceResponse{Balance: 0, Error: err.Error()}, nil
		}
		return nil, err
	}
	return &proto.BalanceResponse{Balance: balance}, nil
}

func (s *UserWalletServer) TopUp(ctx context.Context, req *proto.TopUpRequest) (*proto.TopUpResponse, error) {
	err := s.userWallet.UpdateBalance(ctx, req.Amount)
	if err != nil {
		return nil, err
	}
	return &proto.TopUpResponse{Success: true}, nil
}
