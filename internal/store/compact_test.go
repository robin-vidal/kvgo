package store

import (
	"path/filepath"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func TestStore_SetTriggersCompaction(t *testing.T) {
	dir := t.TempDir()

	db, err := database.New(&config.Config{ShardAmount: 2})
	if err != nil {
		t.Fatalf("database.New() error = %v", err)
	}

	w, err := wal.Open(&config.WalConfig{
		WalPath:             filepath.Join(dir, "wal.log"),
		SnapshotPath:        filepath.Join(dir, "snapshot.db"),
		CompactionThreshold: 3,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer w.Close()

	s := New(db, w)

	if err := s.Set("a", "1"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := s.Set("b", "2"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if snap, _ := w.LoadSnapshot(); snap != nil {
		t.Errorf("snapshot exists before threshold reached: %v", snap)
	}

	if err := s.Set("c", "3"); err != nil { // seqNum = 3, triggers compaction
		t.Fatalf("Set() error = %v", err)
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
		t.Fatalf("Replay() after compaction = %d entries, want 0", len(remaining))
	}
}
