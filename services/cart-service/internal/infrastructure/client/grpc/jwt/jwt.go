package jwt

import (
	"context"
	"fmt"
	"log"

	sso "github.com/vsespontanno/eCommerce/proto/sso"
	"github.com/vsespontanno/eCommerce/services/cart-service/internal/infrastructure/client/grpc/jwt/dto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client sso.ValidatorClient
	addr   string
}

func NewJwtClient(addr string) *Client {
	// addr уже содержит полный адрес: "sso-service.ecommerce.svc.cluster.local:50051"
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial gRPC server %s: %v", addr, err)
	}

	client := sso.NewValidatorClient(conn)
	log.Printf("JWT client connected to %s", addr)

	return &Client{
		client: client,
		addr:   addr,
	}
}

func (j *Client) ValidateToken(ctx context.Context, token string) (*dto.TokenResponse, error) {
	resp, err := j.client.ValidateToken(ctx, &sso.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, fmt.Errorf("validate token failed (addr: %s): %w", j.addr, err)
	}

	return &dto.TokenResponse{Valid: resp.Valid, UserID: resp.UserId}, nil
}
