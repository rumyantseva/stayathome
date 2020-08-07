package internal

import (
	"net"
	"net/http"

	muxtrace "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux"
	oteltrace "go.opentelemetry.io/otel/api/trace"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Diagnostics responsible for diagnostics logic of the app
func Diagnostics(logger *zap.SugaredLogger, tracer oteltrace.Tracer, port string, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()

	mw := muxtrace.Middleware("bl", muxtrace.WithTracer(tracer))
	r.Use(mw)

	r.HandleFunc("/health", handleHealth(logger.With("handler", "health")))

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

func handleHealth(logger *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call")
		w.WriteHeader(http.StatusOK)
	}
}
