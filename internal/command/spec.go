package command

import (
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/wal"
)

type result struct {
	Response []byte
	CmdName  string
	Status   string
}

type arg struct {
	name     string
	kind     string
	optional bool
}

type spec struct {
	name    string
	parent  string
	summary string
	arity   int
	args    []arg
	handler func(db *database.Database, w *wal.WAL, args []string) result
	apply   func(db *database.Database, e wal.Entry) error
}
