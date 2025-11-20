package app

import (
	"github.com/vsespontanno/eCommerce/products-service/internal/app/grpcapp"
	"github.com/vsespontanno/eCommerce/products-service/internal/app/httpapp"
	postgres "github.com/vsespontanno/eCommerce/products-service/internal/infrastructure/repository"
	"github.com/vsespontanno/eCommerce/products-service/internal/presentation/grpc/saga"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	GRPCApp *grpcapp.App
	Store   *postgres.ProductStore
}

func New(logger *zap.SugaredLogger, httpPort int, grpcProductsPort int, grpcSagaPort int, store *postgres.ProductStore, reserver saga.Reserver) *App {
	httpApp := httpapp.New(httpPort, logger)
	grpcApp := grpcapp.NewApp(logger, store, reserver, grpcProductsPort, grpcSagaPort)
	return &App{
		HTTPApp: httpApp,
		GRPCApp: grpcApp,
		Store:   store,
	}
}
