package app

import (
	"database/sql"
	"log/slog"
	"time"

	grpcapp "github.com/vsespontanno/eCommerce/sso-service/internal/app/grpc"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository/postgres"
	"github.com/vsespontanno/eCommerce/sso-service/internal/services/auth"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, db *sql.DB, tokenTTL time.Duration) *App {
	pg := postgres.NewStorage(db)

	authService := auth.NewAuth(log, pg, pg, tokenTTL)

	grpcApp := grpcapp.NewApp(log, authService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
