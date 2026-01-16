package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vsespontanno/eCommerce/pkg/logger"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/app"
	"github.com/vsespontanno/eCommerce/services/wallet-service/internal/config"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}
	logger.InitLogger()

	a, err := app.New(logger.Log, cfg)
	if err != nil {
		logger.Log.Fatalf("Failed to initialize app: %v", err)
	}

	// Start application in goroutine
	go func() {
		a.MustRun()
	}()

	// Wait for interrupt signal for graceful shutdown
	logger.Log.Info("Wallet service started, waiting for shutdown signal")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	logger.Log.Info("Shutdown signal received, initiating graceful shutdown")

	// Graceful shutdown with timeout
	done := make(chan struct{})
	go func() {
		a.Stop()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("Graceful shutdown completed successfully")
	case <-time.After(15 * time.Second):
		logger.Log.Warn("Graceful shutdown timeout exceeded, forcing exit")
	}
}
