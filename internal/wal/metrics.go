package wal

import (
	"context"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metrics struct {
	appendsTotal       metric.Int64Counter
	appendDuration     metric.Float64Histogram
	compactionsTotal   metric.Int64Counter
	compactionDuration metric.Float64Histogram
	replayDuration     metric.Float64Histogram
	replayEntries      metric.Int64Counter
	sizeBytes          metric.Int64ObservableGauge
	snapshotSizeBytes  metric.Int64ObservableGauge
	seqNum             metric.Int64ObservableGauge
}

func newMetrics(wal *WAL) (*metrics, error) {
	meter := otel.Meter("kvgo/wal")

	appendsTotal, err := meter.Int64Counter(
		"wal.appends.total",
		metric.WithDescription("Total number of entries appended to the WAL"),
		metric.WithUnit("{entry}"),
	)
	if err != nil {
		return nil, err
	}

	appendDuration, err := meter.Float64Histogram(
		"wal.append.duration_us",
		metric.WithDescription("Duration of a WAL append, including fsync, in microseconds"),
		metric.WithUnit("us"),
	)
	if err != nil {
		return nil, err
	}

	compactionsTotal, err := meter.Int64Counter(
		"wal.compactions.total",
		metric.WithDescription("Total number of WAL compactions"),
		metric.WithUnit("{compaction}"),
	)
	if err != nil {
		return nil, err
	}

	compactionDuration, err := meter.Float64Histogram(
		"wal.compaction.duration_us",
		metric.WithDescription("Duration of a WAL compaction in microseconds"),
		metric.WithUnit("us"),
	)
	if err != nil {
		return nil, err
	}

	replayDuration, err := meter.Float64Histogram(
		"wal.replay.duration_us",
		metric.WithDescription("Duration of the WAL replay on startup in microseconds"),
		metric.WithUnit("us"),
	)
	if err != nil {
		return nil, err
	}

	replayEntries, err := meter.Int64Counter(
		"wal.replay.entries.total",
		metric.WithDescription("Total number of entries replayed from the WAL on startup"),
		metric.WithUnit("{entry}"),
	)
	if err != nil {
		return nil, err
	}

	sizeBytes, err := meter.Int64ObservableGauge(
		"wal.size",
		metric.WithDescription("Size of the WAL file on disk"),
		metric.WithUnit("By"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			if size, ok := fileSize(wal.cfg.WalPath); ok {
				o.Observe(size)
			}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	snapshotSizeBytes, err := meter.Int64ObservableGauge(
		"wal.snapshot.size",
		metric.WithDescription("Size of the snapshot file on disk"),
		metric.WithUnit("By"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			if size, ok := fileSize(wal.cfg.SnapshotPath); ok {
				o.Observe(size)
			}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	seqNum, err := meter.Int64ObservableGauge(
		"wal.seq_num",
		metric.WithDescription("Sequence number of the last appended WAL entry"),
		metric.WithUnit("{entry}"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			o.Observe(int64(wal.CurrentSeqNum()))
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return &metrics{
		appendsTotal:       appendsTotal,
		appendDuration:     appendDuration,
		compactionsTotal:   compactionsTotal,
		compactionDuration: compactionDuration,
		replayDuration:     replayDuration,
		replayEntries:      replayEntries,
		sizeBytes:          sizeBytes,
		snapshotSizeBytes:  snapshotSizeBytes,
		seqNum:             seqNum,
	}, nil
}

func fileSize(path string) (int64, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, false
	}
	return info.Size(), true
}

func status(err error) string {
	if err != nil {
		return "err"
	}
	return "ok"
}

func (m *metrics) recordAppend(duration time.Duration, err error) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(attribute.String("status", status(err)))
	m.appendsTotal.Add(context.Background(), 1, attrs)
	m.appendDuration.Record(context.Background(), float64(duration.Microseconds()), attrs)
}

func (m *metrics) recordCompaction(duration time.Duration, err error) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(attribute.String("status", status(err)))
	m.compactionsTotal.Add(context.Background(), 1, attrs)
	m.compactionDuration.Record(context.Background(), float64(duration.Microseconds()), attrs)
}

func (m *metrics) recordReplay(duration time.Duration, entries int) {
	if m == nil {
		return
	}
	m.replayDuration.Record(context.Background(), float64(duration.Microseconds()))
	m.replayEntries.Add(context.Background(), int64(entries))
}
