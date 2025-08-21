package app

import (
	"database/sql"

	"github.com/vsespontanno/eCommerce/products-service/internal/app/httpapp"
	"github.com/vsespontanno/eCommerce/products-service/internal/repository/postgres"
	"go.uber.org/zap"
)

type App struct {
	HTTPApp *httpapp.App
	Store   *postgres.ProductStore
}

func New(logger zap.Logger, httpPort int, db *sql.DB) *App {
	store := postgres.NewProductStore(db)
	httpApp := httpapp.New(httpPort, &logger)
	return &App{
		HTTPApp: httpApp,
		Store:   store,
	}
}
