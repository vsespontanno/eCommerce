package app

import (
	"database/sql"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/app/grpcapp"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/repository/postgres"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *zap.SugaredLogger, gRPCServer int, gRPCClient string, db *sql.DB) *App {
	pgWalletUser := postgres.NewWalletUserStore(db)
	pgWalletCreator := postgres.NewWalletCreatorStore(db)
	grpcApp := grpcapp.NewApp(log, pgWalletUser, pgWalletCreator, gRPCServer, gRPCClient)

	return &App{
		GRPCServer: grpcApp,
	}
}
