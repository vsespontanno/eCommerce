package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/app"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/config"
	"github.com/vsespontanno/eCommerce/services/sso-service/internal/repository"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}

	logger.InitLogger()

	// Connect to database with proper error handling
	sDb, err := repository.ConnectToPostgres(cfg.User, cfg.Password, cfg.Name, cfg.Host, cfg.PGPort, logger.Log)
	if err != nil {
		logger.Log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sDb.Close()

	application := app.New(logger.Log, cfg.GRPCPort, cfg.HTTPPort, sDb, cfg.JWTSecret, 1*time.Hour)

	// Start gRPC server
	go func() {
		logger.Log.Info("Starting gRPC server...")
		application.GRPCServer.MustRun()
	}()

	// Start HTTP Gateway
	go func() {
		logger.Log.Info("Starting HTTP Gateway...")
		application.HTTPGateway.MustRun()
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Wait for interrupt signal
	<-stop
	logger.Log.Info("Shutting down servers...")

	// Graceful stop with timeout
	shutdownDone := make(chan struct{})
	go func() {
		application.GRPCServer.Stop()
		application.HTTPGateway.Stop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		logger.Log.Info("Servers gracefully stopped")
	case <-time.After(10 * time.Second):
		logger.Log.Warn("Shutdown timeout exceeded, forcing stop")
	}

	logger.Log.Info("Application stopped")
}
