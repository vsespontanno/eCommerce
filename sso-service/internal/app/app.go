package app

import (
	"database/sql"
	"time"

	grpcapp "github.com/vsespontanno/eCommerce/sso-service/internal/app/grpc"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository/postgres"
	"github.com/vsespontanno/eCommerce/sso-service/internal/services/auth"
	"github.com/vsespontanno/eCommerce/sso-service/internal/services/validator"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *zap.SugaredLogger, grpcPort int, db *sql.DB, jwtSecret string, tokenTTL time.Duration) *App {
	pg := postgres.NewStorage(db)

	authService := auth.NewAuth(log, pg, tokenTTL, jwtSecret)

	validateService := validator.New(jwtSecret)

	grpcApp := grpcapp.NewApp(log, authService, validateService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
