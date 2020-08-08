package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	otelg "go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"

	"github.com/rumyantseva/stayathome/internal"
)

const servicename = "stayathome"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	appLoger := logger.Sugar().Named(servicename)
	appLoger.Info("The application is starting...")

	// Let's choose an exporter, the simplest one just prints everything to stdout:
	exporter, err := stdout.NewExporter(stdout.WithPrettyPrint())
	if err != nil {
		appLoger.Fatalw("Can't enable Open Telemetry exporter", "err", err)
	}

	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		appLoger.Fatal("Jaeger endpoint is not set")
	}
	jExporter, err := jaeger.NewRawExporter(
		jaeger.WithAgentEndpoint(jaegerEndpoint),
		jaeger.WithProcess(jaeger.Process{ServiceName: servicename}),
	)
	if err != nil {
		appLoger.Fatalw("Can't set Jaeger exporter", "err", err)
	}

	// We need to register a global provider first.
	// We use the "AlwaysSample" sampler for debug purposes,
	// but it'll be to slow to keep it for production.
	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(
			sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()},
		),
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSyncer(jExporter),
	)
	if err != nil {
		appLoger.Fatalw("Can't set Open Telemetry provider", "err", err)
	}
	otelg.SetTraceProvider(tp)

	// Now we finally can make a tracer instance to track active spans:
	tracer := otelg.Tracer(servicename)

	// Similar story for metrics
	mc := push.New(
		simple.NewWithExactDistribution(),
		exporter,
	)
	mc.Start()
	defer mc.Stop()
	otelg.SetMeterProvider(mc.Provider())

	pExporter, err := prometheus.NewExportPipeline(prometheus.Config{})
	if err != nil {
		appLoger.Fatalw("Can't set Prometheus exporter", "err", err)
	}
	otelg.SetMeterProvider(pExporter.Provider())

	meter := otelg.Meter(servicename)

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
	bl := internal.BusinessLogic(appLoger.With("module", "bl"), tracer, meter, port, shutdown)
	diag := internal.Diagnostics(appLoger.With("module", "diag"), tracer, pExporter.ServeHTTP, diagPort, shutdown)
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

	err = bl.Shutdown(timeout)
	if err != nil {
		appLoger.Errorw("Got an error from the business logic server", "err", err)
	}
	err = diag.Shutdown(timeout)
	if err != nil {
		appLoger.Errorw("Got an error from the diagnostics server", "err", err)
	}

	appLoger.Info("The application is stopped.")
}
