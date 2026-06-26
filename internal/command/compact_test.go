package command

import (
	"path/filepath"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func TestMaybeCompact(t *testing.T) {
	dir := t.TempDir()
	db := generateSampleDB()

	w, err := wal.Open(&config.WalConfig{
		WalPath:             filepath.Join(dir, "wal.log"),
		SnapshotPath:        filepath.Join(dir, "snapshot.db"),
		CompactionThreshold: 3,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer w.Close()

	Dispatch(db, w, "SET", []string{"a", "1"})
	Dispatch(db, w, "SET", []string{"b", "2"})

	if err := maybeCompact(db, w); err != nil {
		t.Fatalf("maybeCompact() error = %v", err)
	}
	if snap, _ := w.LoadSnapshot(); snap != nil {
		t.Errorf("snapshot exists before threshold reached: %v", snap)
	}

	Dispatch(db, w, "SET", []string{"c", "3"}) // seqNum = 3

	if err := maybeCompact(db, w); err != nil {
		t.Fatalf("maybeCompact() error = %v", err)
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
		t.Fatalf("Replay() after maybeCompact() = %d entries, want 0", len(remaining))
	}
}
