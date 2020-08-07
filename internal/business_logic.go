package internal

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func BusinessLogic(logger *zap.SugaredLogger, port string, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/rent", handleRent(logger.With("handler", "rent"), "http://127.0.0.1:"+port+"/check"))
	r.HandleFunc("/check", handleCheck(logger.With("handler", "check")))

	server := http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: r,
	}

	logger.Info("Ready to start the server...")
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			shutdown <- err
		}
	}()

	return &server
}

func handleRent(logger *zap.SugaredLogger, checkURL string) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call")

		checkr, err := http.Get(checkURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(checkr.StatusCode)
	}
}

func handleCheck(logger *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call")
		w.WriteHeader(http.StatusOK)
	}
}
