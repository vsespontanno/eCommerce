package saga

import (
	"context"

	proto "github.com/vsespontanno/eCommerce/proto/wallet"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/domain/wallet/entity/apperrors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WalletSagaServer struct {
	proto.UnimplementedWalletServer
	sagaWallet SagaWallet
	logger     *zap.SugaredLogger
}

type SagaWallet interface {
	Reserve(ctx context.Context, userID int64, amount int64) error
	Release(ctx context.Context, userID int64, amount int64) error
	Commit(ctx context.Context, userID int64, amount int64) error
}

func NewWalletSagaServer(gRPCServer *grpc.Server, sagaWallet SagaWallet, logger *zap.SugaredLogger) {
	logger.Infow("Registering WalletSaga gRPC server")
	proto.RegisterWalletServer(gRPCServer,
		&WalletSagaServer{
			sagaWallet: sagaWallet,
			logger:     logger,
		},
	)
}

func (s *WalletSagaServer) ReserveFunds(ctx context.Context, req *proto.ReserveFundsRequest) (*proto.ReserveFundsResponse, error) {
	err := s.sagaWallet.Reserve(ctx, req.UserId, req.Amount)
	if err != nil {
		s.logger.Errorw("ReserveFunds failed",
			"userID", req.UserId,
			"amount", req.Amount,
			"error", err,
		)

		switch err {
		case apperrors.ErrNoWallet:
			return nil, status.Error(codes.NotFound, "wallet not found")
		case apperrors.ErrInsufficientFunds:
			return nil, status.Error(codes.FailedPrecondition, "insufficient funds")
		default:
			return nil, status.Error(codes.Internal, "failed to reserve funds")
		}
	}

	s.logger.Infow("ReserveFunds success",
		"userID", req.UserId,
		"amount", req.Amount,
	)

	return &proto.ReserveFundsResponse{Success: true}, nil
}

func (s *WalletSagaServer) ReleaseFunds(ctx context.Context, req *proto.ReleaseFundsRequest) (*proto.ReleaseFundsResponse, error) {
	err := s.sagaWallet.Release(ctx, req.UserId, req.Amount)
	if err != nil {
		s.logger.Errorw("ReleaseFunds failed",
			"userID", req.UserId,
			"amount", req.Amount,
			"error", err,
		)

		switch err {
		case apperrors.ErrNoWallet:
			return nil, status.Error(codes.NotFound, "wallet not found")
		default:
			return nil, status.Error(codes.Internal, "failed to release funds")
		}
	}

	s.logger.Infow("ReleaseFunds success",
		"userID", req.UserId,
		"amount", req.Amount,
	)

	return &proto.ReleaseFundsResponse{Success: true}, nil
}

func (s *WalletSagaServer) CommitFunds(ctx context.Context, req *proto.CommitFundsRequest) (*proto.CommitFundsResponse, error) {
	err := s.sagaWallet.Commit(ctx, req.UserId, req.Amount)
	if err != nil {
		s.logger.Errorw("CommitFunds failed",
			"userID", req.UserId,
			"amount", req.Amount,
			"error", err,
		)

		switch err {
		case apperrors.ErrNoWallet:
			return nil, status.Error(codes.NotFound, "wallet not found")
		default:
			return nil, status.Error(codes.Internal, "failed to commit funds")
		}
	}

	s.logger.Infow("CommitFunds success",
		"userID", req.UserId,
		"amount", req.Amount,
	)

	return &proto.CommitFundsResponse{Success: true}, nil
}
