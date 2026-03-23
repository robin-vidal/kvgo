package telemetry

import (
	"context"
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
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	return provider.Shutdown, nil
}
