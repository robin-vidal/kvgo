package command

import (
	"errors"

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
		handler: func(db *database.Database, args []string) result {
			db.Set(args[0], args[1])
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
