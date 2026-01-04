package httpapp

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcEndpoint := fmt.Sprintf("127.0.0.1:%d", g.grpcPort)

	g.log.Infow("Connecting to gRPC server", "endpoint", grpcEndpoint)

	err := proto.RegisterAuthHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("%s: failed to register auth handler: %w", op, err)
	}

	err = proto.RegisterValidatorHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return fmt.Errorf("%s: failed to register validator handler: %w", op, err)
	}

	httpMux := http.NewServeMux()

	httpMux.Handle("/", mux)

	httpMux.HandleFunc("/health", g.healthCheck)

	g.httpServer = &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", g.port),
		Handler:           httpMux,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	g.log.Infow("HTTP gateway started", "addr", g.httpServer.Addr, "grpc_endpoint", grpcEndpoint)

	if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *Gateway) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"sso-service","timestamp":%d}`, time.Now().Unix())
}

func (g *Gateway) Stop() {
	const op = "httpapp.Stop"

	g.log.Infow("stopping HTTP gateway", "port", g.port, "op", op)

	if err := g.httpServer.Shutdown(context.Background()); err != nil {
		g.log.Errorw("failed to shutdown HTTP server", "op", op, "error", err)
	}
}
