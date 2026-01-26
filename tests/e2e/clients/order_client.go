package clients

import (
	"context"
	"fmt"

	pb "github.com/vsespontanno/eCommerce/proto/orders"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	grpcClient pb.OrderClient
	conn       *grpc.ClientConn
}

func NewOrderClient(address string) (*OrderClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	client := pb.NewOrderClient(conn)
	return &OrderClient{
		grpcClient: client,
		conn:       conn,
	}, nil
}

func (c *OrderClient) Close() error {
	return c.conn.Close()
}

func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*pb.OrderEvent, error) {
	req := &pb.GetOrderRequest{
		OrderId: orderID,
	}

	resp, err := c.grpcClient.GetOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Order, nil
}
