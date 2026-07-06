package command

import (
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/store"
)

func init() {
	defaultRegistry.register(&spec{
		name:    "GET",
		summary: "Get the value of a key.",
		arity:   1,
		args: []arg{
			{name: "key", kind: "key"},
		},
		handler: func(s *store.Store, args []string) result {
			val, ok := s.Get(args[0])
			if !ok {
				return result{Response: resp.EncodeNullBulkString(), Status: "miss"}
			}
			return result{Response: resp.EncodeBulkString(val), Status: "ok"}
		},
	})
}
