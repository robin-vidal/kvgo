package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func Init() (shutdown func(context.Context) error, err error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("kvgo"),
	)

	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(exporter),
	)

	otel.SetMeterProvider(provider)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metrics server stopped", "error", err)
		}
	}()

	return provider.Shutdown, nil
}
