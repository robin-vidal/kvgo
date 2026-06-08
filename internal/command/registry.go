package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
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

func Dispatch(db *database.Database, name string, args []string) result {
	entry, found := defaultRegistry.commands[name]
	if !found {
		return result{
			Response: resp.EncodeError("unknown command " + name),
			CmdName:  name,
			Status:   "err",
		}
	}

	s := entry.spec
	if len(entry.subs) > 0 && len(args) > 0 {
		if subS, found := entry.subs[args[0]]; found {
			s = subS
			args = args[1:]
			name = name + " " + s.name
		} else {
			return result{
				Response: resp.EncodeError("unknown subcommand " + args[0]),
				CmdName:  name,
				Status:   "err",
			}
		}
	}

	if s.arity != len(args) && s.arity != -1 {
		return result{
			Response: resp.EncodeError("wrong number of arguments for '" + s.name + "'"),
			CmdName:  name,
			Status:   "err",
		}
	}

	res := s.handler(db, args)
	res.CmdName = name
	return res
}
