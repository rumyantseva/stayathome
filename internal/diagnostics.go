package internal

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// Diagnostics responsible for diagnostics logic of the app
func Diagnostics(port string, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/health", handleHealth())

	server := http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: r,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			shutdown <- err
		}
	}()

	return &server
}

func handleHealth() func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
