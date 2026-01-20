package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/vsespontanno/eCommerce/pkg/logger"
	proto "github.com/vsespontanno/eCommerce/proto/orders"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/application/order"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/config"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/infrastructure/db"
	"github.com/vsespontanno/eCommerce/services/order-service/internal/infrastructure/repository"
	orderServ "github.com/vsespontanno/eCommerce/services/order-service/internal/presentation/order"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}
	logger.InitLogger()
	db, err := db.ConnectToPostgres(cfg.PGUser, cfg.PGPassword, cfg.PGName, cfg.PGHost, cfg.PGPort)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}
	orderRepo := repository.NewOrderStore(db, logger.Log)
	orderSvc := order.NewOrderService(orderRepo, logger.Log)
	orderServer := orderServ.NewGRPCServer(orderSvc, logger.Log)
	grpcServer := initializeGRPC(logger.Log)

	proto.RegisterOrderServer(grpcServer, orderServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCServerPort))
	if err != nil {
		logger.Log.Fatalf("failed to listen: %v", err)
	}

	logger.Log.Infof("Order gRPC-Server started on port %d", cfg.GRPCServerPort)

	// Start HTTP health check server
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"service": "order-service",
			"time":    time.Now().Unix(),
		}); err != nil {
			logger.Log.Errorw("Failed to encode health response", "error", err)
		}
	})

	healthServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPHealthPort),
		Handler:           healthMux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		logger.Log.Infof("Health check HTTP server started on port %d", cfg.HTTPHealthPort)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Errorw("Health check server stopped", "error", err)
		}
	}()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Log.Errorw("gRPC server stopped", "error", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Log.Info("Shutting down Order service...")

	// Останавливаем gRPC с timeout
	grpcStopDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopDone)
	}()

	select {
	case <-grpcStopDone:
		logger.Log.Info("gRPC server stopped gracefully")
	case <-time.After(10 * time.Second):
		logger.Log.Warn("gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	}

	// Останавливаем HTTP health server
	healthCtx, healthCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer healthCancel()
	if err := healthServer.Shutdown(healthCtx); err != nil {
		logger.Log.Errorw("Health server shutdown failed", "error", err)
	}

	logger.Log.Info("Order service stopped gracefully")
}

func interceptorLogger(l *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		level := zapcore.Level(int8(lvl)) // #nosec G115 - logging.Level range matches zapcore.Level
		l.Log(level, msg)
	})
}

func initializeGRPC(log *zap.SugaredLogger) *grpc.Server {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Errorw("Recovered from panic", "panic", p)
			return status.Errorf(codes.Internal, "internal error")
		}),
	}
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}
	interceptor := interceptorLogger(log)

	return grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptor, loggingOpts...),
	))
}
