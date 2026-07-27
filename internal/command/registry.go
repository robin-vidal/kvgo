package command

import (
	"errors"
	"strings"

	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/store"
	"github.com/robin-vidal/kvgo/internal/wal"
)

var defaultRegistry = &registry{commands: make(map[string]*entry)}

type entry struct {
	spec *spec
	subs map[string]*spec
}

type registry struct {
	commands map[string]*entry
}

func (r *registry) count() int {
	count := 0
	for _, cmd := range r.commands {
		count += 1
		for range cmd.subs {
			count += 1
		}
	}

	return count
}

func (r *registry) register(s *spec) {
	if s == nil {
		panic("no spec provided")
	}

	if s.parent == "" {
		if _, found := r.commands[s.name]; found {
			panic("command already registered: " + s.name)
		}

		r.commands[s.name] = &entry{spec: s, subs: make(map[string]*spec)}
	} else {
		if _, found := r.commands[s.parent]; !found {
			panic("parent command not found: " + s.parent)
		}

		if _, found := r.commands[s.parent].subs[s.name]; found {
			panic("subcommand already registered: " + s.parent + " " + s.name)
		}

		r.commands[s.parent].subs[s.name] = s
	}
}

func Dispatch(st *store.Store, name string, args []string) result {
	name = strings.ToUpper(name)
	entry, found := defaultRegistry.commands[name]
	if !found {
		return result{
			Response: resp.EncodeError("unknown command " + name),
			CmdName:  name,
			Status:   "err",
		}
	}

	spec := entry.spec
	if len(entry.subs) > 0 && len(args) > 0 {
		if subS, found := entry.subs[args[0]]; found {
			spec = subS
			args = args[1:]
			name = name + " " + spec.name
		} else {
			return result{
				Response: resp.EncodeError("unknown subcommand " + args[0]),
				CmdName:  name,
				Status:   "err",
			}
		}
	}

	if spec.arity != len(args) && spec.arity != -1 {
		return result{
			Response: resp.EncodeError("wrong number of arguments for '" + spec.name + "'"),
			CmdName:  name,
			Status:   "err",
		}
	}

	res := spec.handler(st, args)
	res.CmdName = name
	return res
}

func Apply(db *database.Database, e wal.Entry) error {
	cmd, found := defaultRegistry.commands[e.Command]

	if !found || cmd.spec.apply == nil {
		return errors.New("command cannot be applied '" + e.Command + "'")
	}

	return cmd.spec.apply(db, e)
}
