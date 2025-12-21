package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/vsespontanno/eCommerce/proto/wallet"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Gateway represents HTTP gateway for gRPC services
type Gateway struct {
	mux        *runtime.ServeMux
	grpcAddr   string
	httpPort   int
	logger     *zap.SugaredLogger
	httpServer *http.Server
}

// NewGateway creates a new HTTP gateway
func NewGateway(grpcAddr string, httpPort int, logger *zap.SugaredLogger) *Gateway {
	// Create gRPC-Gateway mux with custom options
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(customErrorHandler),
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
	)

	return &Gateway{
		mux:      mux,
		grpcAddr: grpcAddr,
		httpPort: httpPort,
		logger:   logger,
	}
}

// Start starts the HTTP gateway server
func (g *Gateway) Start(ctx context.Context) error {
	// Setup gRPC connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Register WalletTopUP service (user-facing REST API)
	err := wallet.RegisterWalletTopUPHandlerFromEndpoint(ctx, g.mux, g.grpcAddr, opts)
	if err != nil {
		return fmt.Errorf("failed to register WalletTopUP handler: %w", err)
	}

	g.logger.Infow("Registered WalletTopUP HTTP handlers",
		"grpcAddr", g.grpcAddr,
	)

	// Create HTTP server
	g.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", g.httpPort),
		Handler:      corsMiddleware(loggingMiddleware(g.mux, g.logger)),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
		IdleTimeout:  1 * time.Second,
	}

	g.logger.Infow("Starting HTTP gateway server",
		"port", g.httpPort,
		"grpcAddr", g.grpcAddr,
	)

	// Start server
	if err := g.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP gateway: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP gateway server
func (g *Gateway) Stop(ctx context.Context) error {
	if g.httpServer == nil {
		return nil
	}

	g.logger.Info("Stopping HTTP gateway server")
	if err := g.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP gateway: %w", err)
	}

	g.logger.Info("HTTP gateway server stopped successfully")
	return nil
}

// customErrorHandler handles errors from gRPC-Gateway
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

// customHeaderMatcher matches incoming HTTP headers to gRPC metadata
func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization", "authorization":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infow("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
		)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
