package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/store"
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
		handler: func(s *store.Store, args []string) result {
			if err := s.Delete(args[0]); err != nil {
				return result{Response: resp.EncodeError(err.Error()), Status: "err"}
			}
			return result{Response: resp.EncodeInteger(1), Status: "ok"}
		},
		apply: func(db *database.Database, e wal.Entry) error {
			db.Delete(e.Key)
			return nil
		},
	})
}
