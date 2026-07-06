package command

import (
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/store"
)

func init() {
	defaultRegistry.register(&spec{
		name:    "PING",
		summary: "Ping the server.",
		arity:   0,
		handler: func(s *store.Store, args []string) result {
			return result{Response: resp.EncodeSimpleString("PONG"), Status: "ok"}
		},
	})
}
