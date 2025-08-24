package ssowallet

import (
	"context"
	"log"

	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"google.golang.org/grpc"
)

type SsoWalletServer struct {
	proto.UnimplementedCreatorServer
	walletCreator WalletCreator
}

type WalletCreator interface {
	CreateWallet(ctx context.Context, userID int64) (bool, string, error)
}

func NewSsoWalletServer(gRPCServer *grpc.Server, walletCreator WalletCreator) {
	log.Println("Registering SsoWalletServer")
	proto.RegisterCreatorServer(gRPCServer, &SsoWalletServer{walletCreator: walletCreator})
}

func (s *SsoWalletServer) CreateWallet(ctx context.Context, req *proto.CreateWalletRequest) (*proto.CreateWalletResponse, error) {
	success, msg, err := s.walletCreator.CreateWallet(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &proto.CreateWalletResponse{Success: success, Message: msg}, nil
}
