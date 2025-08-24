package client

import (
	"context"
	"log"

	wallet "github.com/vsespontanno/eCommerce/proto/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type WalletClient struct {
	client wallet.CreatorClient
	port   string
}

func NewWalletClient(port string) *WalletClient {
	addr := "localhost:" + port
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}
	client := wallet.NewCreatorClient(conn)
	return &WalletClient{
		client: client,
		port:   port,
	}
}
func (j *WalletClient) CreateWallet(ctx context.Context, userID int64) (bool, string, error) {
	resp, err := j.client.CreateWallet(ctx, &wallet.CreateWalletRequest{UserId: userID})
	if err != nil {
		return false, "", err
	}
	return resp.Success, resp.Message, nil
}
