package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery" // only for the logging.Level conversion helper
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/application/saga"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/application/user"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/config"
	db "github.com/vsespontanno/eCommerce/services/wallet-service/internal/infrastructure/db/postgres"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/infrastructure/grpcClient"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/infrastructure/repository/postgres"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/presentation/gateway"
	sagaServ "github.com/vsespontanno/eCommerce/services/wallet-service/internal/presentation/server/saga"
	userServ "github.com/vsespontanno/eCommerce/services/wallet-service/internal/presentation/server/user"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/presentation/server/user/interceptor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// App holds servers, database connection, gateway and logger
type App struct {
	Log         *zap.SugaredLogger
	usrServ     *grpc.Server
	sagaSrv     *grpc.Server
	gateway     *gateway.Gateway
	db          *sql.DB
	userPort    int
	sagaPort    int
	gatewayPort int
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

	// Initialize HTTP Gateway
	grpcUserAddr := fmt.Sprintf("localhost:%d", cfg.GRPCUserServer)
	gw := gateway.NewGateway(grpcUserAddr, cfg.HTTPGateway, logger)

	return &App{
		Log:         logger,
		usrServ:     userGRPCServer,
		sagaSrv:     sagaGRPCServer,
		gateway:     gw,
		db:          dataBase,
		userPort:    cfg.GRPCUserServer,
		sagaPort:    cfg.GRPCSagaServer,
		gatewayPort: cfg.HTTPGateway,
	}, nil
}

// initializeGRPC creates a grpc.Server with unified interceptor chain.
// if authInterceptor != nil it is appended as the last interceptor (so it runs after logging/recovery).
func initializeGRPC(log *zap.SugaredLogger, authInterceptor grpc.UnaryServerInterceptor) *grpc.Server {
	// recovery handler with stacktrace
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
		authInterceptor,
	))

	return gRPCServer
}
func interceptorLogger(l *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		level := zapcore.Level(lvl)
		l.Log(level, msg)
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

	errCh := make(chan error, 3)

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

	// start HTTP gateway
	go func() {
		a.Log.Infof("HTTP Gateway starting on :%d", a.gatewayPort)
		if err := a.gateway.Start(ctx); err != nil {
			a.Log.Infow("HTTP gateway stopped", "error", err)
		}
	}()

	// wait for either context cancel (signal) or any server error
	select {
	case <-ctx.Done():
		// shutdown initiated by signal
		a.Log.Infow("shutdown signal received, graceful stopping servers", "reason", ctx.Err())
		st := time.Now()

		// Stop HTTP gateway
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.gateway.Stop(shutdownCtx); err != nil {
			a.Log.Errorw("failed to stop HTTP gateway", "error", err)
		}

		// Stop gRPC servers
		a.usrServ.GracefulStop()
		a.sagaSrv.GracefulStop()

		a.Log.Infow("servers stopped", "duration", time.Since(st).String())
		return nil
	case err := <-errCh:
		// server returned an unexpected error
		a.Log.Errorw("server returned error", "error", err)

		// attempt graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.gateway.Stop(shutdownCtx)

		a.usrServ.GracefulStop()
		a.sagaSrv.GracefulStop()
		return err
	}
}

// Stop attempts graceful stop (can be called externally)
func (a *App) Stop() {
	a.Log.Info("shutting down all servers and database connection")

	// Stop HTTP gateway
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.gateway.Stop(ctx); err != nil {
		a.Log.Errorw("failed to stop HTTP gateway", "error", err)
	}

	// Stop gRPC servers
	a.usrServ.GracefulStop()
	a.sagaSrv.GracefulStop()

	// Close database connection
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.Log.Errorw("failed to close database connection", "error", err)
		} else {
			a.Log.Info("database connection closed successfully")
		}
	}
}
