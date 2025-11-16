package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery" // only for the logging.Level conversion helper
	"github.com/vsespontanno/eCommerce/wallet-service/internal/application/saga"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/application/user"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/config"
	db "github.com/vsespontanno/eCommerce/wallet-service/internal/infrastructure/db/postgres"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/infrastructure/grpcClient"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/infrastructure/repository/postgres"
	sagaServ "github.com/vsespontanno/eCommerce/wallet-service/internal/presentation/server/saga"
	userServ "github.com/vsespontanno/eCommerce/wallet-service/internal/presentation/server/user"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/presentation/server/user/interceptor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// App holds servers and logger
type App struct {
	Log      *zap.SugaredLogger
	usrServ  *grpc.Server
	sagaSrv  *grpc.Server
	userPort int
	sagaPort int
}

// New builds the app (does not run servers)
func New(logger *zap.SugaredLogger, cfg *config.Config) (*App, error) {
	// connect to DB
	dataBase, err := db.ConnectToPostgres(cfg.User, cfg.Password, cfg.Name, cfg.Host, cfg.PGPort)
	if err != nil {
		// proper sugared logging without %w
		logger.Errorw("failed to connect to Postgres", "error", err)
		return nil, err
	}

	usrRepo := postgres.NewWalletUserStore(dataBase)
	sagaRepo := postgres.NewSagaWalletStore(dataBase)
	gRPCClient := grpcClient.NewJwtClient(cfg.GRPCClient)

	// initialize servers with interceptor chain
	authInt := interceptor.AuthInterceptor()
	// pass auth interceptor only to user server
	userGRPCServer := initializeGRPC(logger, authInt)
	// saga server doesn't need auth interceptor (example) but chain is otherwise identical
	sagaGRPCServer := initializeGRPC(logger, nil)

	userSvc := user.NewWalletService(usrRepo, gRPCClient, logger)
	sagaSvc := saga.NewSagaWalletService(sagaRepo, logger)

	userServ.NewUserWalletServer(userGRPCServer, userSvc, logger)
	sagaServ.NewWalletSagaServer(sagaGRPCServer, sagaSvc, logger)

	return &App{
		Log:      logger,
		usrServ:  userGRPCServer,
		sagaSrv:  sagaGRPCServer,
		userPort: cfg.GRPCUserServer,
		sagaPort: cfg.GRPCSagaServer,
	}, nil
}

// initializeGRPC creates a grpc.Server with unified interceptor chain.
// if authInterceptor != nil it is appended as the last interceptor (so it runs after logging/recovery).
func initializeGRPC(log *zap.SugaredLogger, authInterceptor grpc.UnaryServerInterceptor) *grpc.Server {
	// recovery handler with stacktrace
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			// log panic and stacktrace
			log.Errorw("panic recovered in grpc handler", "panic", p, "stack", string(debug.Stack()))
			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	// we will log payload received/sent
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}

	// our log adapter that preserves fields
	logAdapter := interceptorLogger(log)

	// build interceptor slice and preserve same order always
	interceptors := []grpc.UnaryServerInterceptor{
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(logAdapter, loggingOpts...),
	}

	if authInterceptor != nil {
		interceptors = append(interceptors, authInterceptor)
	}

	return grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
}

// interceptorLogger converts grpc-middleware logging.Level -> zap level and preserves fields
func interceptorLogger(l *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		// convert to zapcore.Level
		// grpc_logging.Logrus library exposes level conversion helper; we only need numeric mapping.
		// fall back to InfoLevel when unknown.
		var zlevel zapcore.Level
		switch lvl {
		case logging.LevelDebug:
			zlevel = zapcore.DebugLevel
		case logging.LevelInfo:
			zlevel = zapcore.InfoLevel
		case logging.LevelWarn:
			zlevel = zapcore.WarnLevel
		case logging.LevelError:
			zlevel = zapcore.ErrorLevel
		default:
			zlevel = zapcore.InfoLevel
		}
		// include fields if present
		if len(fields) > 0 {
			// SugaredLogger.Log accepts level and variadic args, but expects the pattern: msg, keys/values...
			// We'll call Log with msg and fields
			l.Log(zlevel, append([]any{msg}, fields...)...)
			return
		}
		l.Log(zlevel, msg)
	})
}

// MustRun panics on error (convenience wrapper)
func (a *App) MustRun() {
	a.Log.Info("running grpc servers (must)")
	if err := a.Run(); err != nil {
		// panic so supervisor notices
		panic(err)
	}
}

// Run starts both gRPC servers and waits until an error occurs or termination signal received.
func (a *App) Run() error {
	const op = "grpcapp.Run"

	// create cancellable context that listens to SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 2)

	// start user server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.userPort))
		if err != nil {
			errCh <- fmt.Errorf("%s (user listen): %w", op, err)
			return
		}
		a.Log.Infof("User gRPC server listening on :%d", a.userPort)
		if err := a.usrServ.Serve(lis); err != nil {
			// ignore normal server stop
			if err == grpc.ErrServerStopped {
				a.Log.Infow("user grpc server stopped gracefully")
				return
			}
			errCh <- fmt.Errorf("%s (user serve): %w", op, err)
		}
	}()

	// start saga server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.sagaPort))
		if err != nil {
			errCh <- fmt.Errorf("%s (saga listen): %w", op, err)
			return
		}
		a.Log.Infof("Saga gRPC server listening on :%d", a.sagaPort)
		if err := a.sagaSrv.Serve(lis); err != nil {
			if err == grpc.ErrServerStopped {
				a.Log.Infow("saga grpc server stopped gracefully")
				return
			}
			errCh <- fmt.Errorf("%s (saga serve): %w", op, err)
		}
	}()

	// wait for either context cancel (signal) or any server error
	select {
	case <-ctx.Done():
		// shutdown initiated by signal
		a.Log.Infow("shutdown signal received, graceful stopping servers", "reason", ctx.Err())
		st := time.Now()
		a.usrServ.GracefulStop()
		a.sagaSrv.GracefulStop()
		a.Log.Infow("servers stopped", "duration", time.Since(st).String())
		return nil
	case err := <-errCh:
		// server returned an unexpected error
		a.Log.Errorw("grpc server returned error", "error", err)
		// attempt graceful shutdown
		a.usrServ.GracefulStop()
		a.sagaSrv.GracefulStop()
		return err
	}
}

// Stop attempts graceful stop (can be called externally)
func (a *App) Stop() {
	a.Log.Info("shutting down the grpc servers (Stop called)")
	a.usrServ.GracefulStop()
	a.sagaSrv.GracefulStop()
}
