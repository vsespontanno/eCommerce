package app

import (
	"fmt"
	"net"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/application/user"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/config"
	db "github.com/vsespontanno/eCommerce/wallet-service/internal/infrastructure/db/postgres"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/infrastructure/grpcClient"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/infrastructure/repository/postgres"
	userServ "github.com/vsespontanno/eCommerce/wallet-service/internal/presentation/server/user"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/presentation/server/user/interceptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	Log      *zap.SugaredLogger
	usrServ  *grpc.Server
	userPort int
}

func New(logger *zap.SugaredLogger, cfg *config.Config) (*App, error) {
	dataBase, err := db.ConnectToPostgres(cfg.User, cfg.Password, cfg.Name, cfg.Host, cfg.PGPort)
	if err != nil {
		logger.Errorf("Error while connecting to db: %w", err)
		return nil, err
	}
	usrRepo := postgres.NewWalletUserStore(dataBase)
	gRPCClient := grpcClient.NewJwtClient(cfg.GRPCClient)
	gRPCServerWallet := grpc.NewServer(grpc.UnaryInterceptor(interceptor.AuthInterceptor()))
	userSvc := user.NewWalletService(usrRepo, gRPCClient, logger)
	userServ.NewUserWalletServer(gRPCServerWallet, userSvc)
	return &App{
		Log:      logger,
		usrServ:  gRPCServerWallet,
		userPort: cfg.GRPCServer,
	}, nil
}

func (a *App) MustRun() {
	a.Log.Info("running grpc server")
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.userPort))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := a.usrServ.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	a.Log.Info("the user grpc server is running on port %v", a.userPort)

	return nil
}

func (a *App) Stop() {
	a.Log.Info("shutting down the grpc servers")
	a.usrServ.GracefulStop()
}
