package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Log.Errorw("gRPC server stopped", "error", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Log.Info("Shutting down Saga orchestrator...")
	grpcServer.GracefulStop()
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
