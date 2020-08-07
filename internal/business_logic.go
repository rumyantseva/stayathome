package internal

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	muxtrace "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux"
	oteltrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
	"go.uber.org/zap"
)

func BusinessLogic(logger *zap.SugaredLogger, tracer oteltrace.Tracer, port string, shutdown chan<- error) *http.Server {
	r := mux.NewRouter()

	mw := muxtrace.Middleware("bl", muxtrace.WithTracer(tracer))
	r.Use(mw)

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

		req, err := http.NewRequest(http.MethodGet, checkURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Errorw("Error when creating request to check", "err", err)
			return
		}

		httptrace.Inject(r.Context(), req)
		checkr, err := http.DefaultClient.Do(req)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Errorw("Error when sending request to check", "err", err)
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
