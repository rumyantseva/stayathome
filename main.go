package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar().Named("grahovac")
	sugar.Info("The application is starting...")

	port := os.Getenv("PORT")
	if port == "" {
		sugar.Fatal("PORT is not set")
	}

	diagPort := os.Getenv("DIAG_PORT")
	if diagPort == "" {
		sugar.Fatal("DIAG_PORT is not set")
	}

	r := mux.NewRouter()
	server := http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: r,
	}

	diagLogger := sugar.With("subapp", "diag_router")
	diagRouter := mux.NewRouter()
	diagRouter.HandleFunc("/health", func(
		w http.ResponseWriter, _ *http.Request) {
		diagLogger.Info("Health was called")
		w.WriteHeader(http.StatusOK)
	})

	diag := http.Server{
		Addr:    net.JoinHostPort("", diagPort),
		Handler: diagRouter,
	}

	shutdown := make(chan error, 2)

	sugar.Infof("Business logic server is starting...")
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			shutdown <- err
		}
	}()

	sugar.Infof("Diagnostics server is starting...")
	go func() {
		err := diag.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			shutdown <- err
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case x := <-interrupt:
		sugar.Infow("Received", "signal", x.String())

	case err := <-shutdown:
		sugar.Errorw("Error from functional unit", "err", err)
	}

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err := server.Shutdown(timeout)
	if err != nil {
		sugar.Errorw("The business logic is stopped with error", "err", err)
	}

	err = diag.Shutdown(timeout)
	if err != nil {
		sugar.Errorw("The diagnostics server is stopped with error", "err", err)
	}

	sugar.Info("The application is stopped.")
}
