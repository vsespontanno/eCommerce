package grpcapp

import (
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/vsespontanno/eCommerce/products-service/internal/grpc/products"
	"github.com/vsespontanno/eCommerce/products-service/internal/grpc/saga"
	proto "github.com/vsespontanno/eCommerce/proto/products"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	log            *zap.SugaredLogger
	productsServer *grpc.Server
	sagaServer     *grpc.Server
	productsPort   int
	sagaPort       int
}

func NewApp(log *zap.SugaredLogger, productsInterface products.Products, sagaReserver saga.Reserver, productsPort, sagaPort int) *App {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Errorw("Recovered from panic", "panic", p)
			return status.Errorf(codes.Internal, "internal error")
		}),
	}
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}
	interceptorLogger := interceptorLogger(log)

	productsServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLogger, loggingOpts...),
	))
	sagaServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLogger, loggingOpts...),
	))

	// Register servers
	products.NewProductServer(productsServer, productsInterface)
	proto.RegisterSagaProductsServer(sagaServer, saga.NewSagaServer(sagaReserver))

	return &App{
		log:            log,
		productsServer: productsServer,
		sagaServer:     sagaServer,
		productsPort:   productsPort,
		sagaPort:       sagaPort,
	}
}

func interceptorLogger(l *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		level := zapcore.Level(lvl)
		l.Log(level, msg)
	})
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	productsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.productsPort))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	sagaListener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.sagaPort))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		a.log.Infow("products gRPC server started", "port", a.productsPort)
		if err := a.productsServer.Serve(productsListener); err != nil {
			a.log.Errorw("products server failed", "error", err)
		}
	}()

	a.log.Infow("saga gRPC server started", "port", a.sagaPort)
	if err := a.sagaServer.Serve(sagaListener); err != nil {
		return fmt.Errorf("%s: saga server failed: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	a.log.Infow("stopping products gRPC server", "port", a.productsPort)
	a.productsServer.GracefulStop()

	a.log.Infow("stopping saga gRPC server", "port", a.sagaPort)
	a.sagaServer.GracefulStop()
}
