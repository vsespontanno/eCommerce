package client

import (
	"context"
	"log"

	sso "github.com/vsespontanno/eCommerce/proto/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type JwtClient struct {
	client sso.ValidatorClient
	port   string
}

func NewJwtClient(port string) *JwtClient {
	addr := "localhost:" + port
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}
	client := sso.NewValidatorClient(conn)
	return &JwtClient{
		client: client,
		port:   port,
	}
}
func (j *JwtClient) ValidateToken(ctx context.Context, token string) (*sso.ValidateTokenResponse, error) {
	resp, err := j.client.ValidateToken(ctx, &sso.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
