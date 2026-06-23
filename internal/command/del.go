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
		handler: func(db *database.Database, args []string) result {
			db.Delete(args[0])
			return result{Response: resp.EncodeInteger(1), Status: "ok"}
		},
		apply: func(db *database.Database, e wal.Entry) error {
			db.Delete(e.Key)
			return nil
		},
	})
}
