package saga

import (
	"context"

	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type WalletSagaServer struct {
	sagaWallet SagaWallet
	logger     *zap.SugaredLogger
	proto.UnimplementedWalletServer
}

type SagaWallet interface {
	Reserve(ctx context.Context, userID int64, amount int64) error
	Release(ctx context.Context, userID int64, amount int64) error
	Commit(ctx context.Context, userID int64, amount int64) error
}

func NewWalletSagaServer(gRPCServer *grpc.Server, sagaWallet SagaWallet, logger *zap.SugaredLogger) {
	logger.Info("Registering wallet service")
	proto.RegisterWalletServer(gRPCServer, &WalletSagaServer{sagaWallet: sagaWallet, logger: logger})

}

func (s *WalletSagaServer) ReserveFunds(ctx context.Context, req *proto.ReserveFundsRequest) (*proto.ReserveFundsResponse, error) {
	err := s.sagaWallet.Reserve(ctx, req.UserId, req.Amount)
	if err != nil {
		s.logger.Errorw("error while reserving funds", "error", err, "stage", "WalletSagaServer.ReserveFunds")
		return nil, err
	}
	return &proto.ReserveFundsResponse{Success: true}, nil
}

func (s *WalletSagaServer) ReleaseFunds(ctx context.Context, req *proto.ReleaseFundsRequest) (*proto.ReleaseFundsResponse, error) {
	err := s.sagaWallet.Release(ctx, req.UserId, req.Amount)
	if err != nil {
		s.logger.Errorw("error while releasing funds", "error", err, "stage", "WalletSagaServer.ReleaseFunds")
		return nil, err
	}
	return &proto.ReleaseFundsResponse{Success: true}, nil
}

func (s *WalletSagaServer) CommitFunds(ctx context.Context, req *proto.CommitFundsRequest) (*proto.CommitFundsResponse, error) {
	err := s.sagaWallet.Commit(ctx, req.UserId, req.Amount)
	if err != nil {
		s.logger.Errorw("error while committing funds", "error", err, "stage", "WalletSagaServer.CommitFunds")
		return nil, err
	}
	return &proto.CommitFundsResponse{Success: true}, nil
}
