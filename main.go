package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rumyantseva/stayathome/internal"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	appLoger := logger.Sugar().Named("stayathome")
	appLoger.Info("The application is starting...")

	appLoger.Info("Reading configuration...")
	port := os.Getenv("PORT")
	if port == "" {
		appLoger.Fatal("PORT is not set")
	}
	diagPort := os.Getenv("DIAG_PORT")
	if diagPort == "" {
		appLoger.Fatal("DIAG_PORT is not set")
	}
	appLoger.Info("Configuration is ready")

	shutdown := make(chan error, 2)
	bl := internal.BusinessLogic(appLoger.With("module", "bl"), port, shutdown)
	diag := internal.Diagnostics(appLoger.With("module", "diag"), diagPort, shutdown)
	appLoger.Info("Servers are ready")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case x := <-interrupt:
		appLoger.Infow("Received", "signal", x.String())
	case err := <-shutdown:
		appLoger.Errorw("Received error from functional unit", "err", err)
	}

	appLoger.Info("Stopping the servers...")
	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err := bl.Shutdown(timeout)
	if err != nil {
		appLoger.Errorw("Got an error from the business logic server", "err", err)
	}
	err = diag.Shutdown(timeout)
	if err != nil {
		appLoger.Errorw("Got an error from the diagnostics server", "err", err)
	}

	appLoger.Info("The application is stopped.")
}
