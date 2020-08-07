package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rumyantseva/stayathome/internal"
)

func main() {
	port := os.Getenv("PORT")
	diagPort := os.Getenv("DIAG_PORT")

	shutdown := make(chan error, 2)
	bl := internal.BusinessLogic(port, shutdown)
	diag := internal.Diagnostics(diagPort, shutdown)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	select {
	case <-interrupt:
	case <-shutdown:
	}

	timeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	bl.Shutdown(timeout)
	diag.Shutdown(timeout)
}
