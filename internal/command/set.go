package command

import (
	"errors"
	"log/slog"

	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func init() {
	defaultRegistry.register(&spec{
		name:    "SET",
		summary: "Set the string value of a key.",
		arity:   2,
		args: []arg{
			{name: "key", kind: "key"},
			{name: "value", kind: "string"},
		},
		handler: func(db *database.Database, w *wal.WAL, args []string) result {
			entry := wal.Entry{
				Command: "SET",
				Key:     args[0],
				Value:   &args[1],
			}

			err := w.Append(entry)
			if err != nil {
				return result{Response: resp.EncodeError(err.Error()), Status: "err"}
			}

			db.Set(args[0], args[1])

			if w.ShouldCompact() {
				if w.ShouldCompact() {
					if err := maybeCompact(db, w); err != nil {
						slog.Error("compaction failed", "error", err)
					}
				}
			}

			return result{Response: resp.EncodeSimpleString("OK"), Status: "ok"}
		},
		apply: func(db *database.Database, e wal.Entry) error {
			if e.Value == nil {
				return errors.New("SET entry has nil value")
			}
			db.Set(e.Key, *e.Value)

			return nil
		},
	})
}
