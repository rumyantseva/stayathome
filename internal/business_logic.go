package internal

import (
	"context"
	"net"
	"net/http"

	"go.opentelemetry.io/otel/api/metric"

	"github.com/gorilla/mux"
	muxtrace "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux"
	oteltrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
	"go.uber.org/zap"
)

func BusinessLogic(logger *zap.SugaredLogger, tracer oteltrace.Tracer, meter metric.Meter, port string, shutdown chan<- error) *http.Server {
	rentCounter := metric.Must(meter).NewInt64Counter("rent.count")

	r := mux.NewRouter()

	mw := muxtrace.Middleware("bl", muxtrace.WithTracer(tracer))
	r.Use(mw)

	r.HandleFunc("/rent", handleRent(logger.With("handler", "rent"), rentCounter, "http://127.0.0.1:"+port+"/check"))
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

func handleRent(logger *zap.SugaredLogger, counter metric.Int64Counter, checkURL string) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call")

		req, err := http.NewRequest(http.MethodGet, checkURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		httptrace.Inject(r.Context(), req)
		checkr, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		counter.Add(context.Background(), 1)
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
