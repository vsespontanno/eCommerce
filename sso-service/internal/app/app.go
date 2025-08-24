package app

import (
	"database/sql"
	"log/slog"
	"time"

	grpcapp "github.com/vsespontanno/eCommerce/sso-service/internal/app/grpc"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository/postgres"
	"github.com/vsespontanno/eCommerce/sso-service/internal/services/auth"
	"github.com/vsespontanno/eCommerce/sso-service/internal/services/validator"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, clientPort string, db *sql.DB, jwtSecret string, tokenTTL time.Duration) *App {
	pg := postgres.NewStorage(db)

	authService := auth.NewAuth(log, pg, tokenTTL, jwtSecret)

	validateService := validator.New(jwtSecret)

	grpcApp := grpcapp.NewApp(log, authService, validateService, grpcPort, clientPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
