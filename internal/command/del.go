package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func init() {
	defaultRegistry.register(&spec{
		name:    "DEL",
		summary: "Delete a key.",
		arity:   1,
		args: []arg{
			{name: "key", kind: "key"},
		},
		handler: func(db *database.Database, w *wal.WAL, args []string) result {
			entry := wal.Entry{
				Command: "DEL",
				Key:     args[0],
			}

			err := w.Append(entry)
			if err != nil {
				return result{Response: resp.EncodeError(err.Error()), Status: "err"}
			}

			db.Delete(args[0])

			if w.ShouldCompact() {
				w.Compact(db.Dump())
			}

			return result{Response: resp.EncodeInteger(1), Status: "ok"}
		},
		apply: func(db *database.Database, e wal.Entry) error {
			db.Delete(e.Key)
			return nil
		},
	})
}
