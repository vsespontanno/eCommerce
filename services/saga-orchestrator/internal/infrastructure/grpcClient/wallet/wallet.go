package wallet

import (
	"context"
	"fmt"
	"log"

	"github.com/vsespontanno/eCommerce/proto/wallet"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const Success = "success"

type Client struct {
	client wallet.WalletClient
	logger *zap.SugaredLogger
	addr   string
}

func NewWalletClient(addr string, logger *zap.SugaredLogger) *Client {
	// addr уже содержит полный адрес из ConfigMap
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}

	client := wallet.NewWalletClient(conn)
	logger.Infow("Connected to Wallet service", "addr", addr)
	return &Client{
		client: client,
		addr:   addr,
		logger: logger,
	}
}

func (w *Client) ReserveFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	response, err := w.client.ReserveFunds(ctx, &wallet.ReserveFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error reserving funds", "error", err, "userID", userID, "amount", amount)
		return "", err
	}
	if !response.Success {
		w.logger.Errorw("Failed to reserve funds", "error", response.Message, "userID", userID, "amount", amount)
		return "", fmt.Errorf("reserve funds failed: %s", response.Message)
	}
	w.logger.Infow("Funds reserved successfully", "userID", userID, "amount", amount)
	return Success, nil
}

func (w *Client) CommitFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	resp, err := w.client.CommitFunds(ctx, &wallet.CommitFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error committing funds", "error", err, "userID", userID, "amount", amount)
		return "", err
	}
	if !resp.Success {
		w.logger.Errorw("Failed to commit funds", "error", resp.Message, "userID", userID, "amount", amount)
		return "", fmt.Errorf("commit funds failed: %s", resp.Message)
	}
	w.logger.Infow("Funds committed successfully", "userID", userID, "amount", amount)
	return Success, nil
}

func (w *Client) ReleaseFunds(ctx context.Context, userID int64, amount int64) (string, error) {
	resp, err := w.client.ReleaseFunds(ctx, &wallet.ReleaseFundsRequest{
		UserId: userID,
		Amount: amount,
	})
	if err != nil {
		w.logger.Errorw("Error releasing funds", "error", err, "userID", userID, "amount", amount)
		return "", err
	}
	if !resp.Success {
		w.logger.Errorw("Failed to release funds", "error", resp.Message, "userID", userID, "amount", amount)
		return "", fmt.Errorf("release funds failed: %s", resp.Message)
	}
	w.logger.Infow("Funds released successfully", "userID", userID, "amount", amount)
	return Success, nil
}
