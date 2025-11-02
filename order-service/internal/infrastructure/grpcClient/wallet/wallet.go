package wallet

import (
	"context"
	"log"

	"github.com/vsespontanno/eCommerce/proto/wallet"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WalletClient struct {
	client wallet.WalletClient
	logger *zap.SugaredLogger
	port   string
}

func NewWalletClient(port string) WalletClient {
	address := "localhost:" + port

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", address, err)
	}

	client := wallet.NewWalletClient(conn)
	return WalletClient{
		client: client,
		port:   port,
	}
}

func (w *WalletClient) ReserveFunds(ctx context.Context, userID int64, amount int64) error {
	_, err := w.client.ReserveFunds(ctx, &wallet.ReserveFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error reserving funds", "error", err)
		return err
	}
	return err
}

func (w *WalletClient) CommitFunds(ctx context.Context, userID int64, amount int64) error {
	_, err := w.client.CommitFunds(ctx, &wallet.CommitFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error committing funds", "error", err)
		return err
	}
	return err
}

func (w *WalletClient) ReleaseFunds(ctx context.Context, userID int64, amount int64) error {
	_, err := w.client.ReleaseFunds(ctx, &wallet.ReleaseFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error releasing funds", "error", err)
		return err
	}
	return err
}
