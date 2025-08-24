package grpcapp

import (
	"fmt"
	"net"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/client"
	creator "github.com/vsespontanno/eCommerce/wallet-service/internal/grpc/sso-wallet"
	topup "github.com/vsespontanno/eCommerce/wallet-service/internal/grpc/wallet-user"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/grpc/wallet-user/interceptor"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	log        *zap.SugaredLogger
	gRPCServer *grpc.Server
	ServerPort int
	ClientPort string
}

func NewApp(log *zap.SugaredLogger, userWallet topup.UserWallet, walletCreator creator.WalletCreator, serverPort int, clientPort string) *App {
	gRPCClient := client.NewJwtClient(clientPort)
	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.AuthInterceptor(gRPCClient)))

	topup.NewUserWalletServer(gRPCServer, userWallet)
	creator.NewSsoWalletServer(gRPCServer, walletCreator)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		ServerPort: serverPort,
		ClientPort: clientPort,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.ServerPort))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Infof("grpc server started, addr: %s", l.Addr().String())
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.Infof("stopping gRPC server, op: %s, port: %d", op, a.ServerPort)

	a.gRPCServer.GracefulStop()
}
