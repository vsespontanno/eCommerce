package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/vsespontanno/eCommerce/pkg/logger"
	proto "github.com/vsespontanno/eCommerce/proto/saga"
	applicationSaga "github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/application/saga"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/config"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/infrastructure/db"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/infrastructure/grpcClient/products"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/infrastructure/grpcClient/wallet"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/infrastructure/outbox"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/infrastructure/repository"
	"github.com/vsespontanno/eCommerce/services/saga-orchestrator/internal/presentation/saga"
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

	// PostgreSQL init
	postgresDB, err := db.NewPostgresDB(cfg, logger.Log)
	if err != nil {
		logger.Log.Fatalw("failed to connect to postgres", "error", err)
	}
	defer postgresDB.Close()

	// Outbox repository
	outboxRepo := repository.NewOutboxRepository(postgresDB, logger.Log)

	// Kafka producer для outbox publisher
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.KafkaBroker,
		"acks":               "all",
		"retries":            10,
		"enable.idempotence": true,
	})
	if err != nil {
		logger.Log.Fatalw("failed to create kafka producer", "error", err)
	}
	defer kafkaProducer.Close()

	// Outbox publisher (фоновый worker)
	outboxPublisher := outbox.NewOutboxPublisher(
		postgresDB,
		kafkaProducer,
		logger.Log,
		cfg.KafkaTopic,
		5*time.Second,
	)

	// Запускаем outbox publisher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	outboxPublisher.Start(ctx)

	// gRPC clients
	walletClient := wallet.NewWalletClient(cfg.GRPCWalletClientPort, logger.Log)
	productsClient := products.NewProductsClient(cfg.GRPCProductsClientPort, logger.Log)

	// Saga service (использует outbox)
	sagaService := applicationSaga.New(cfg, walletClient, productsClient, outboxRepo, logger.Log)
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

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Log.Info("Shutting down Saga orchestrator...")

	// Останавливаем outbox publisher
	cancel()
	time.Sleep(1 * time.Second)

	// Останавливаем gRPC
	grpcServer.GracefulStop()

	logger.Log.Info("Saga orchestrator stopped gracefully")
}

func interceptorLogger(l *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		// #nosec G115 - logging.Level и zapcore.Level имеют одинаковые значения
		level := zapcore.Level(int8(lvl))
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
