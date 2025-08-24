package wallettopup

import (
	"context"
	"log"

	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/grpc/wallet-user/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserWalletServer struct {
	proto.UnimplementedWalletTopUPServer
	userWallet UserWallet
}

type UserWallet interface {
	GetBalance(ctx context.Context, userID int64) (balance float64, err error)
	UpdateBalance(ctx context.Context, userID int64, amount float64) (err error)
}

func NewUserWalletServer(gRPCServer *grpc.Server, userWallet UserWallet) {
	log.Println("Registering UserWalletServer")
	proto.RegisterWalletTopUPServer(gRPCServer, &UserWalletServer{userWallet: userWallet})
}

func (s *UserWalletServer) Balance(ctx context.Context, req *proto.BalanceRequest) (*proto.BalanceResponse, error) {
	userID, ok := ctx.Value(interceptor.UserIDKey).(int64)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}
	balance, err := s.userWallet.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &proto.BalanceResponse{Balance: balance}, nil
}

func (s *UserWalletServer) TopUp(ctx context.Context, req *proto.TopUpRequest) (*proto.TopUpResponse, error) {
	userID, ok := ctx.Value(interceptor.UserIDKey).(int64)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user ID")
	}
	err := s.userWallet.UpdateBalance(ctx, userID, req.Amount)
	if err != nil {
		return nil, err
	}
	return &proto.TopUpResponse{Success: true}, nil
}
