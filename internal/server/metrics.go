package server

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metrics struct {
	commandsTotal     metric.Int64Counter
	commandsDuration  metric.Float64Histogram
	connectionsActive metric.Int64UpDownCounter
}

func newMetrics() (*metrics, error) {
	meter := otel.Meter("kvgo/server")

	commandsTotal, err := meter.Int64Counter(
		"db.commands.total",
		metric.WithDescription("Total number of commands executed"),
		metric.WithUnit("{command}"),
	)
	if err != nil {
		return nil, err
	}

	commandsDuration, err := meter.Float64Histogram(
		"db.commands.duration_us",
		metric.WithDescription("Duration of command execution in microseconds"),
		metric.WithUnit("us"),
	)
	if err != nil {
		return nil, err
	}

	connectionsActive, err := meter.Int64UpDownCounter(
		"db.connections.active",
		metric.WithDescription("Number of active TCP connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return nil, err
	}

	return &metrics{
		commandsTotal:     commandsTotal,
		commandsDuration:  commandsDuration,
		connectionsActive: connectionsActive,
	}, nil
}

// recordCommand increments the commands counter for the given command and status.
func (m *metrics) recordCommand(cmd, status string) {
	m.commandsTotal.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("command", cmd),
			attribute.String("status", status),
		),
	)
}

// recordDuration records the execution duration of a command in milliseconds.
func (m *metrics) recordDuration(cmd string, duration float64) {
	m.commandsDuration.Record(context.Background(), duration,
		metric.WithAttributes(
			attribute.String("command", cmd),
		),
	)
}
