package internal

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

func BusinessLogic(port string, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/rent", handleRent("http://127.0.0.1:"+port+"/check"))
	r.HandleFunc("/check", handleCheck())

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

func handleRent(checkURL string) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {

		checkr, err := http.Get(checkURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(checkr.StatusCode)
	}
}

func handleCheck() func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
