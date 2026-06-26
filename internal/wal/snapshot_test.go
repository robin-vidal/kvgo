package wal

import (
	"path/filepath"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
)

func TestSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.db")
	w := &WAL{cfg: &config.WalConfig{SnapshotPath: path}}

	data := map[string]string{"foo": "bar", "baz": "qux"}
	if err := w.Snapshot(data); err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	got, err := w.LoadSnapshot()
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v", err)
	}
	if len(got) != len(data) {
		t.Fatalf("LoadSnapshot() = %d keys, want %d", len(got), len(data))
	}
	for k, v := range data {
		if got[k] != v {
			t.Errorf("LoadSnapshot()[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestLoadSnapshot_NoExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	w := &WAL{cfg: &config.WalConfig{SnapshotPath: path}}

	got, err := w.LoadSnapshot()
	if err != nil || got != nil {
		t.Fatalf("LoadSnapshot() = %v, %v, want nil, nil", got, err)
	}
}

func TestTruncate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")
	w, err := Open(&config.WalConfig{WalPath: path})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer w.Close()

	for _, k := range []string{"a", "b", "c", "d"} {
		v := "v-" + k
		if err := w.Append(Entry{Command: "SET", Key: k, Value: &v}); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	// drops "a"=1 and "b"=2
	if err := w.truncate(2); err != nil {
		t.Fatalf("truncate() error = %v", err)
	}

	got, err := w.Replay()
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Replay() after truncate = %d entries, want 2", len(got))
	}
	if got[0].Key != "c" || got[1].Key != "d" {
		t.Errorf("Replay() after truncate = %+v, want keys c, d", got)
	}
}

func TestCompact(t *testing.T) {
	dir := t.TempDir()
	w, err := Open(&config.WalConfig{
		WalPath:      filepath.Join(dir, "wal.log"),
		SnapshotPath: filepath.Join(dir, "snapshot.db"),
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer w.Close()

	for _, k := range []string{"a", "b", "c"} {
		v := "v-" + k
		w.Append(Entry{Command: "SET", Key: k, Value: &v})
	}

	snapshotSeqNum := w.CurrentSeqNum()
	data := map[string]string{"a": "v-a", "b": "v-b", "c": "v-c"}

	if err := w.Compact(data, snapshotSeqNum); err != nil {
		t.Fatalf("Compact() error = %v", err)
	}

	snap, err := w.LoadSnapshot()
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v", err)
	}
	if len(snap) != 3 {
		t.Fatalf("LoadSnapshot() = %d keys, want 3", len(snap))
	}

	remaining, err := w.Replay()
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("Replay() after Compact() = %d entries, want 0", len(remaining))
	}
}

func TestShouldCompact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")
	w, err := Open(&config.WalConfig{WalPath: path, CompactionThreshold: 3})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer w.Close()

	v := "v"
	for i := 0; i < 2; i++ {
		w.Append(Entry{Command: "SET", Key: "k", Value: &v})
		if w.ShouldCompact() {
			t.Errorf("ShouldCompact() = true before threshold (seqNum=%d)", w.CurrentSeqNum())
		}
	}
	w.Append(Entry{Command: "SET", Key: "k", Value: &v}) // seqNum = 3
	if !w.ShouldCompact() {
		t.Errorf("ShouldCompact() = false at threshold (seqNum=%d)", w.CurrentSeqNum())
	}
}
