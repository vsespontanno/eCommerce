package app

import (
	"fmt"
	"net"

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
	"google.golang.org/grpc"
)

type App struct {
	Log      *zap.SugaredLogger
	usrServ  *grpc.Server
	sagaSrv  *grpc.Server
	userPort int
	sagaPort int
}

func New(logger *zap.SugaredLogger, cfg *config.Config) (*App, error) {
	dataBase, err := db.ConnectToPostgres(cfg.User, cfg.Password, cfg.Name, cfg.Host, cfg.PGPort)
	if err != nil {
		logger.Errorf("Error while connecting to db: %w", err)
		return nil, err
	}
	usrRepo := postgres.NewWalletUserStore(dataBase)
	sagaRepo := postgres.NewSagaWalletStore(dataBase)
	gRPCClient := grpcClient.NewJwtClient(cfg.GRPCClient)
	gRPCServerWallet := grpc.NewServer(grpc.UnaryInterceptor(interceptor.AuthInterceptor()))
	gRPCServerSaga := grpc.NewServer()

	userSvc := user.NewWalletService(usrRepo, gRPCClient, logger)
	sagaSvc := saga.NewSagaWalletService(sagaRepo, logger)
	userServ.NewUserWalletServer(gRPCServerWallet, userSvc)
	sagaServ.NewWalletSagaServer(gRPCServerSaga, sagaSvc, logger)
	return &App{
		Log:      logger,
		usrServ:  gRPCServerWallet,
		sagaSrv:  gRPCServerSaga,
		userPort: cfg.GRPCUserServer,
		sagaPort: cfg.GRPCSagaServer,
	}, nil
}

func (a *App) MustRun() {
	a.Log.Info("running grpc servers")
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	errCh := make(chan error, 2)

	// user server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.userPort))
		if err != nil {
			errCh <- fmt.Errorf("%s (user): %w", op, err)
			return
		}
		a.Log.Infof("User gRPC server running on port %d", a.userPort)
		if err := a.usrServ.Serve(lis); err != nil {
			errCh <- fmt.Errorf("%s (user serve): %w", op, err)
		}
	}()

	// saga server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.sagaPort))
		if err != nil {
			errCh <- fmt.Errorf("%s (saga): %w", op, err)
			return
		}
		a.Log.Infof("Saga gRPC server running on port %d", a.sagaPort)
		if err := a.sagaSrv.Serve(lis); err != nil {
			errCh <- fmt.Errorf("%s (saga serve): %w", op, err)
		}
	}()

	// если один из серверов вернёт ошибку — завершаем всё приложение
	return <-errCh
}

func (a *App) Stop() {
	a.Log.Info("shutting down the grpc servers")
	a.usrServ.GracefulStop()
	a.sagaSrv.GracefulStop()
}
