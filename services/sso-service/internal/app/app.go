package app

import (
	"database/sql"
	"time"

	grpcapp "github.com/vsespontanno/eCommerce/services/sso-service/internal/app/grpc"
	httpapp "github.com/vsespontanno/eCommerce/services/sso-service/internal/app/http"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/repository/postgres"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/services/auth"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/services/validator"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer  *grpcapp.App
	HTTPGateway *httpapp.Gateway
}

func New(log *zap.SugaredLogger, grpcPort, httpPort int, db *sql.DB, jwtSecret string, tokenTTL time.Duration) *App {
	pg := postgres.NewStorage(db)

	authService := auth.NewAuth(log, pg, tokenTTL, jwtSecret)

	validateService := validator.New(jwtSecret)

	grpcApp := grpcapp.NewApp(log, authService, validateService, grpcPort)

	httpGateway := httpapp.NewGateway(log, httpPort, grpcPort)

	return &App{
		GRPCServer:  grpcApp,
		HTTPGateway: httpGateway,
	}
}
