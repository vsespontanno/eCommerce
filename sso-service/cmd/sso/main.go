package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/vsespontanno/eCommerce/sso-service/internal/app"
	"github.com/vsespontanno/eCommerce/sso-service/internal/config"
	"github.com/vsespontanno/eCommerce/sso-service/internal/repository"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	logger := setupLogger(cfg.Env)
	sDb, err := repository.ConnectToPostgres(cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.Host, cfg.DB.Port)
	if err != nil {
		panic("Failed to connect to db: " + err.Error())
	}
	application := app.New(logger, cfg.GRPC.Port, sDb, cfg.TokenTTL)
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

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
