package grpcClient

import (
	"context"
	"log"
	"strings"

	"github.com/vsespontanno/eCommerce/proto/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type JwtClient struct {
	client sso.ValidatorClient
	port   string
}

func NewJwtClient(port string) *JwtClient {
	// If port contains ':', it's already a full address (e.g., sso-service.ecommerce.svc.cluster.local:50051)
	// Otherwise, it's just a port number and we need to add localhost
	addr := port
	if !strings.Contains(port, ":") {
		addr = "localhost:" + port
	}
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
func (j *JwtClient) ValidateToken(ctx context.Context, token string) (userID int64, err error) {
	resp, err := j.client.ValidateToken(ctx, &sso.ValidateTokenRequest{Token: token})
	if err != nil {
		return 0, err
	}
	if resp.Valid {
		return resp.UserId, nil
	}
	return 0, nil
}
