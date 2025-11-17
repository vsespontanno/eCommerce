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
	proto "github.com/vsespontanno/eCommerce/proto/saga"
	applicationSaga "github.com/vsespontanno/eCommerce/saga-orchestrator/internal/application/saga"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/config"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/infrastructure/grpcClient/products"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/infrastructure/grpcClient/wallet"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/infrastructure/messaging"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/presentation/saga"
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

	// Kafka init
	kafkaProducer, err := messaging.NewKafkaProducer(cfg.KafkaBroker, cfg.KafkaTopic, logger.Log)
	if err != nil {
		logger.Log.Fatalw("failed to create kafka producer", "error", err)
	}
	defer kafkaProducer.Close()
	walletClient := wallet.NewWalletClient(cfg.GRPCWalletClientPort, logger.Log)
	productsClient := products.NewProductsClient(cfg.GRPCProductsClientPort, logger.Log)
	// TODO: inject real clients later (wallet, products)
	sagaService := applicationSaga.New(cfg, &walletClient, &productsClient, kafkaProducer, logger.Log)
	sagaServer := saga.NewSagaServer(logger.Log, sagaService)

	grpcServer := initializeGRPC(logger.Log)
	proto.RegisterSagaServer(grpcServer, sagaServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCServerPort))
	if err != nil {
		logger.Log.Fatalf("failed to listen: %v", err)
	}

	logger.Log.Infof("Saga orchestrator gRPC server started on port %d", cfg.GRPCServerPort)

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
		level := zapcore.Level(lvl)
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
