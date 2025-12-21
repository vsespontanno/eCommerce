package main

import (
	"os"
	"os/signal"
	"syscall"

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
		logger.Log.Fatal(err)
	}
	go func() {
		a.MustRun()
	}()

	// Graceful shutdown
	a.Log.Info("signal sigint")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	a.Stop()
}
