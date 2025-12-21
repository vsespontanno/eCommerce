package httpapp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	proto "github.com/vsespontanno/eCommerce/proto/sso"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	log        *zap.SugaredLogger
	httpServer *http.Server
	port       int
	grpcPort   int
}

func NewGateway(log *zap.SugaredLogger, httpPort, grpcPort int) *Gateway {
	return &Gateway{
		log:      log,
		port:     httpPort,
		grpcPort: grpcPort,
	}
}

func (g *Gateway) MustRun() {
	if err := g.Run(); err != nil {
		panic(err)
	}
}

func (g *Gateway) Run() error {
	const op = "httpapp.Run"

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// Register gRPC-Gateway handlers
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	// Используем 127.0.0.1 вместо localhost для совместимости с Docker
	grpcEndpoint := fmt.Sprintf("127.0.0.1:%d", g.grpcPort)

	g.log.Infow("Connecting to gRPC server", "endpoint", grpcEndpoint)

	// Register Auth service
	err := proto.RegisterAuthHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("%s: failed to register auth handler: %w", op, err)
	}

	// Register Validator service
	err = proto.RegisterValidatorHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("%s: failed to register validator handler: %w", op, err)
	}

	// Create HTTP server
	g.httpServer = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", g.port),
		Handler: mux,
	}

	g.log.Infow("HTTP gateway started", "addr", g.httpServer.Addr, "grpc_endpoint", grpcEndpoint)

	if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *Gateway) Stop() {
	const op = "httpapp.Stop"

	g.log.Infow("stopping HTTP gateway", "port", g.port, "op", op)

	if err := g.httpServer.Shutdown(context.Background()); err != nil {
		g.log.Errorw("failed to shutdown HTTP server", "op", op, "error", err)
	}
}
