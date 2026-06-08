package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
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
	})
}
