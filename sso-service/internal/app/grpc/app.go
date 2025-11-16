package grpcapp

import (
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	authgrpc "github.com/vsespontanno/eCommerce/sso-service/internal/grpc/auth"
	validategrpc "github.com/vsespontanno/eCommerce/sso-service/internal/grpc/validator"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	log        *zap.SugaredLogger
	gRPCServer *grpc.Server
	port       int
}

func NewApp(log *zap.SugaredLogger, authService authgrpc.Auth, validator validategrpc.Validator, port int) *App {
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Errorw("Recovered from panic", "panic", p)
			return status.Errorf(codes.Internal, "internal error")
		}),
	}
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(interceptorLogger(log), loggingOpts...),
	))
	authgrpc.NewAuthServer(gRPCServer, authService)
	validategrpc.NewValidationServer(gRPCServer, validator)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func interceptorLogger(l *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		level := zapcore.Level(lvl)
		l.Log(level, msg)
	})
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("user server started", zap.Stringer("addr", l.Addr()))
	if err := a.gRPCServer.Serve(l); err != nil {
		a.log.Errorw(op, zap.Error(err))
		return err
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.Info("stopping gRPC server",
		zap.Int("port", a.port),
		zap.String("op", op))

	a.gRPCServer.GracefulStop()
}
