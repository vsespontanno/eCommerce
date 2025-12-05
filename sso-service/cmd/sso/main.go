package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/sso-service/internal/app"
	"github.com/vsespontanno/eCommerce/sso-service/internal/config"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}
	logger.InitLogger()
	sDb, err := repository.ConnectToPostgres(cfg.User, cfg.Password, cfg.Name, cfg.Host, cfg.PGPort)
	if err != nil {
		panic("Failed to connect to db: " + err.Error())
	}

	// HTTP port по умолчанию 8080
	httpPort := 8080

	application := app.New(logger.Log, cfg.GRPCPort, httpPort, sDb, cfg.JWTSecret, time.Duration(1*time.Hour))

	// Запуск gRPC сервера
	go func() {
		application.GRPCServer.MustRun()
	}()

	// Даем gRPC серверу время на запуск перед стартом HTTP Gateway
	time.Sleep(1 * time.Second)

	// Запуск HTTP Gateway
	go func() {
		application.HTTPGateway.MustRun()
	}()

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGINT (pkill -2) or SIGTERM
	<-stop

	// initiate graceful shutdown
	application.GRPCServer.Stop()
	application.HTTPGateway.Stop()
	logger.Log.Infow("Gracefully stopped")
}
