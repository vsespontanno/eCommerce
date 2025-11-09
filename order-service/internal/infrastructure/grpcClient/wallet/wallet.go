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

func NewWalletClient(port string, logger *zap.SugaredLogger) WalletClient {
	address := "localhost:" + port

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", address, err)
	}

	client := wallet.NewWalletClient(conn)
	return WalletClient{
		client: client,
		port:   port,
		logger: logger,
	}
}

func (w *WalletClient) ReserveFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	response, err := w.client.ReserveFunds(ctx, &wallet.ReserveFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error reserving funds", "error", err, "stage", "WalletClient.ReserveFunds")
		return response.String(), err
	}
	return response.String(), err
}

func (w *WalletClient) CommitFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	resp, err := w.client.CommitFunds(ctx, &wallet.CommitFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error committing funds", "error", err, "stage", "WalletClient.CommitFunds")
		return resp.String(), err
	}
	return resp.String(), err
}

func (w *WalletClient) ReleaseFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	resp, err := w.client.ReleaseFunds(ctx, &wallet.ReleaseFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error releasing funds", "error", err, "stage", "WalletClient.ReleaseFunds")
		return resp.String(), err
	}
	return resp.String(), err
}
