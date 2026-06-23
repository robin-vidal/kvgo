package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func init() {
	defaultRegistry.register(&spec{
		name:    "PING",
		summary: "Ping the server.",
		arity:   0,
		handler: func(db *database.Database, w *wal.WAL, args []string) result {
			return result{Response: resp.EncodeSimpleString("PONG"), Status: "ok"}
		},
	})
}
