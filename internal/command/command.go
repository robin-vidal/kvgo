package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func init() {
	defaultRegistry.register(&spec{
		name:    "COMMAND",
		summary: "List available commands.",
		arity:   0,
		handler: func(db *database.Database, w *wal.WAL, args []string) result {
			elems := [][]byte{}
			for name, e := range defaultRegistry.commands {
				elems = append(elems, encodeCommandInfo(name, e.spec))
				for subName, sub := range e.subs {
					elems = append(elems, encodeCommandInfo(name+"|"+subName, sub))
				}
			}
			return result{Response: resp.EncodeArray(elems), Status: "ok"}
		},
	})

	defaultRegistry.register(&spec{
		name:    "DOCS",
		parent:  "COMMAND",
		summary: "Return documentation for commands.",
		arity:   0,
		handler: func(db *database.Database, w *wal.WAL, args []string) result {
			elems := [][]byte{}
			for name, e := range defaultRegistry.commands {
				elems = append(elems, resp.EncodeBulkString(name))
				elems = append(elems, encodeCommandDocs(e.spec))
				for subName, sub := range e.subs {
					elems = append(elems, resp.EncodeBulkString(name+"|"+subName))
					elems = append(elems, encodeCommandDocs(sub))
				}
			}
			return result{Response: resp.EncodeArray(elems), Status: "ok"}
		},
	})

	defaultRegistry.register(&spec{
		name:    "COUNT",
		parent:  "COMMAND",
		summary: "Return the number of registered commands.",
		arity:   0,
		handler: func(db *database.Database, w *wal.WAL, args []string) result {
			return result{Response: resp.EncodeInteger(defaultRegistry.count()), Status: "ok"}
		},
	})
}

func encodeCommandInfo(name string, s *spec) []byte {
	return resp.EncodeArray([][]byte{
		resp.EncodeBulkString(name),
		resp.EncodeInteger(s.arity + 1),
		resp.EncodeArray([][]byte{}),
		resp.EncodeInteger(0),
		resp.EncodeInteger(0),
		resp.EncodeInteger(0),
	})
}

func encodeCommandDocs(s *spec) []byte {
	return resp.EncodeArray([][]byte{
		resp.EncodeBulkString("summary"),
		resp.EncodeBulkString(s.summary),
	})
}
