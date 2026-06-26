package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func maybeCompact(db *database.Database, w *wal.WAL) error {
	if !w.ShouldCompact() {
		return nil
	}

	snapshotSeqNum := w.CurrentSeqNum()
	data := db.Dump()

	return w.Compact(data, snapshotSeqNum)
}
