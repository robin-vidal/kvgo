package store

import (
	"log/slog"
	"sync"

	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/wal"
)

type Store struct {
	db      *database.Database
	wal     *wal.WAL
	applyMu sync.Mutex
}

func New(db *database.Database, w *wal.WAL) *Store {
	return &Store{db: db, wal: w}
}

func (s *Store) Get(key string) (string, bool) {
	return s.db.Get(key)
}

func (s *Store) Set(key, value string) error {
	s.applyMu.Lock()
	defer s.applyMu.Unlock()

	entry := wal.Entry{
		Command: "SET",
		Key:     key,
		Value:   &value,
	}

	if err := s.wal.Append(entry); err != nil {
		return err
	}

	s.db.Set(key, value)

	if s.wal.ShouldCompact() {
		if err := s.maybeCompact(); err != nil {
			slog.Error("compaction failed", "error", err)
		}
	}

	return nil
}

func (s *Store) Delete(key string) error {
	s.applyMu.Lock()
	defer s.applyMu.Unlock()

	entry := wal.Entry{
		Command: "DEL",
		Key:     key,
	}

	if err := s.wal.Append(entry); err != nil {
		return err
	}

	s.db.Delete(key)

	if s.wal.ShouldCompact() {
		if err := s.maybeCompact(); err != nil {
			slog.Error("compaction failed", "error", err)
		}
	}

	return nil
}

func (s *Store) maybeCompact() error {
	snapshotSeqNum := s.wal.CurrentSeqNum()
	data := s.db.Dump()

	return s.wal.Compact(data, snapshotSeqNum)
}

func (s *Store) GetKeyAmountPerShard() []int {
	return s.db.GetKeyAmountPerShard()
}
