package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vsespontanno/eCommerce/wallet-service/internal/app"
	"github.com/vsespontanno/eCommerce/wallet-service/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.MustLoad()
	if err != nil {
		panic(err)
	}
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	a, err := app.New(sugar, cfg)
	go func() {
		a.MustRun()
	}()

	// Graceful shutdown
	a.Log.Info("signal sigint")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	a.Stop() // Assuming GRPCServer has Stop() method for graceful shutdown
}
