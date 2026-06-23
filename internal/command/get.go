package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func init() {
	defaultRegistry.register(&spec{
		name:    "GET",
		summary: "Get the value of a key.",
		arity:   1,
		args: []arg{
			{name: "key", kind: "key"},
		},
		handler: func(db *database.Database, w *wal.WAL, args []string) result {
			val, ok := db.Get(args[0])
			if !ok {
				return result{Response: resp.EncodeNullBulkString(), Status: "miss"}
			}
			return result{Response: resp.EncodeBulkString(val), Status: "ok"}
		},
	})
}
