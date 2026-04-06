package server

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/robin-vidal/kvgo/internal/database"
)

type metrics struct {
	commandsTotal     metric.Int64Counter
	commandsDuration  metric.Float64Histogram
	connectionsActive metric.Int64UpDownCounter
	storeKeys         metric.Int64ObservableGauge
	goroutines        metric.Int64ObservableGauge
	heapAlloc         metric.Int64ObservableGauge
}

func newMetrics(db *database.Database) (*metrics, error) {
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

	storeKeys, err := meter.Int64ObservableGauge(
		"db.store.keys",
		metric.WithDescription("Number of keys per shard"),
		metric.WithUnit("{key}"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			for idx, amount := range db.GetKeyAmountPerShard() {
				o.Observe(int64(amount), metric.WithAttributes(attribute.Int("shard", idx)))
			}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	goroutines, err := meter.Int64ObservableGauge(
		"process.goroutines",
		metric.WithDescription("Number of goroutines"),
		metric.WithUnit("{goroutine}"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(runtime.NumGoroutine()))
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	heapAlloc, err := meter.Int64ObservableGauge(
		"process.mem.heap_alloc",
		metric.WithDescription("Amount of memory allocated on the heap"),
		metric.WithUnit("By"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			o.Observe(int64(memStats.HeapAlloc))
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return &metrics{
		commandsTotal:     commandsTotal,
		commandsDuration:  commandsDuration,
		connectionsActive: connectionsActive,
		storeKeys:         storeKeys,
		goroutines:        goroutines,
		heapAlloc:         heapAlloc,
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

// recordConnection increments or decrements the active connections counter.
func (m *metrics) recordConnection(delta int64) {
	m.connectionsActive.Add(context.Background(), delta)
}
