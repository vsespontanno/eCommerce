package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/app"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/config"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/repository"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}
	sDb, err := repository.ConnectToPostgres(cfg.User, cfg.Password, cfg.Name, cfg.Host, cfg.PGPort)
	if err != nil {
		panic("Failed to connect to db: " + err.Error())
	}
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	application := app.New(sugar, cfg.GRPCServer, cfg.GRPCClient, sDb)
	go func() {
		application.GRPCServer.MustRun()
	}()

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Waiting for SIGINT (pkill -2) or SIGTERM
	<-stop

	// initiate graceful shutdown
	application.GRPCServer.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
	logger.Info("Gracefully stopped")
}
