package database

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metrics struct {
	opsTotal metric.Int64Counter
	lockWait metric.Float64Histogram
}

func newMetrics() (*metrics, error) {
	meter := otel.Meter("kvgo/database")

	opsTotal, err := meter.Int64Counter(
		"db.shard.ops.total",
		metric.WithDescription("Total number of operations per shard"),
		metric.WithUnit("{operation}"),
	)
	if err != nil {
		return nil, err
	}

	lockWait, err := meter.Float64Histogram(
		"db.shard.lock_wait_us",
		metric.WithDescription("Time spent waiting to acquire a shard lock in microseconds"),
		metric.WithUnit("us"),
	)
	if err != nil {
		return nil, err
	}

	return &metrics{
		opsTotal: opsTotal,
		lockWait: lockWait,
	}, nil
}

func (m *metrics) record(shard int, op, mode string, waited time.Duration) {
	if m == nil {
		return
	}
	m.opsTotal.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.Int("shard", shard),
			attribute.String("op", op),
		),
	)
	m.lockWait.Record(context.Background(), float64(waited.Microseconds()),
		metric.WithAttributes(
			attribute.Int("shard", shard),
			attribute.String("mode", mode),
		),
	)
}
