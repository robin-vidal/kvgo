package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
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
		handler: func(db *database.Database, args []string) result {
			db.Set(args[0], args[1])
			return result{Response: resp.EncodeSimpleString("OK"), Status: "ok"}
		},
	})
}
