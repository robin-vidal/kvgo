package command

import (
	"bytes"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
)

func dummyHandler(db *database.Database, args []string) result {
	return result{Response: resp.EncodeSimpleString("OK"), Status: "ok"}
}

func generateSampleDB() *database.Database {
	return database.New(&config.Config{
		Host:        "localhost",
		Port:        6379,
		Debug:       false,
		ShardAmount: 2,
	})
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(r *registry)
		spec      *spec
		wantCount int
	}{
		{
			name:      "Top Level Command",
			setup:     func(r *registry) {},
			spec:      &spec{name: "FOO", handler: dummyHandler},
			wantCount: 1,
		},
		{
			name: "Subcommand",
			setup: func(r *registry) {
				r.register(&spec{name: "FOO", handler: dummyHandler})
			},
			spec:      &spec{name: "BAR", parent: "FOO", handler: dummyHandler},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registry{commands: make(map[string]*entry)}
			tt.setup(r)

			r.register(tt.spec)

			if got := r.count(); got != tt.wantCount {
				t.Errorf("count() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

func TestRegisterPanics(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(r *registry)
		spec      *spec
		wantPanic string
	}{
		{
			name:      "Nil Spec",
			setup:     func(r *registry) {},
			spec:      nil,
			wantPanic: "no spec provided",
		},
		{
			name: "Duplicate Command",
			setup: func(r *registry) {
				r.register(&spec{name: "FOO", handler: dummyHandler})
			},
			spec:      &spec{name: "FOO", handler: dummyHandler},
			wantPanic: "command already registered: FOO",
		},
		{
			name:      "Parent Not Found",
			setup:     func(r *registry) {},
			spec:      &spec{name: "BAR", parent: "FOO", handler: dummyHandler},
			wantPanic: "parent command not found: FOO",
		},
		{
			name: "Duplicate Subcommand",
			setup: func(r *registry) {
				r.register(&spec{name: "FOO", handler: dummyHandler})
				r.register(&spec{name: "BAR", parent: "FOO", handler: dummyHandler})
			},
			spec:      &spec{name: "BAR", parent: "FOO", handler: dummyHandler},
			wantPanic: "subcommand already registered: FOO BAR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registry{commands: make(map[string]*entry)}
			tt.setup(r)

			defer func() {
				got := recover()
				if got == nil {
					t.Errorf("register() expected panic, got nil")
					return
				}
				if msg, ok := got.(string); !ok || msg != tt.wantPanic {
					t.Errorf("register() panic = %v, want %v", got, tt.wantPanic)
				}
			}()

			r.register(tt.spec)
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name  string
		setup func(r *registry)
		want  int
	}{
		{
			name:  "Empty",
			setup: func(r *registry) {},
			want:  0,
		},
		{
			name: "Top Level Only",
			setup: func(r *registry) {
				r.register(&spec{name: "FOO", handler: dummyHandler})
				r.register(&spec{name: "BAZ", handler: dummyHandler})
			},
			want: 2,
		},
		{
			name: "With Subcommands",
			setup: func(r *registry) {
				r.register(&spec{name: "FOO", handler: dummyHandler})
				r.register(&spec{name: "BAR", parent: "FOO", handler: dummyHandler})
				r.register(&spec{name: "BAZ", parent: "FOO", handler: dummyHandler})
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &registry{commands: make(map[string]*entry)}
			tt.setup(r)

			if got := r.count(); got != tt.want {
				t.Errorf("count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDispatch(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(db *database.Database)
		cmd         string
		args        []string
		wantResp    []byte
		wantCmdName string
		wantStatus  string
	}{
		{
			name:        "SET",
			setup:       func(db *database.Database) {},
			cmd:         "SET",
			args:        []string{"foo", "bar"},
			wantResp:    resp.EncodeSimpleString("OK"),
			wantCmdName: "SET",
			wantStatus:  "ok",
		},
		{
			name: "GET Hit",
			setup: func(db *database.Database) {
				db.Set("foo", "bar")
			},
			cmd:         "GET",
			args:        []string{"foo"},
			wantResp:    resp.EncodeBulkString("bar"),
			wantCmdName: "GET",
			wantStatus:  "ok",
		},
		{
			name:        "GET Miss",
			setup:       func(db *database.Database) {},
			cmd:         "GET",
			args:        []string{"missing"},
			wantResp:    resp.EncodeNullBulkString(),
			wantCmdName: "GET",
			wantStatus:  "miss",
		},
		{
			name:        "PING",
			setup:       func(db *database.Database) {},
			cmd:         "PING",
			args:        []string{},
			wantResp:    resp.EncodeSimpleString("PONG"),
			wantCmdName: "PING",
			wantStatus:  "ok",
		},
		{
			name:        "Unknown Command",
			setup:       func(db *database.Database) {},
			cmd:         "UNKNOWN",
			args:        []string{},
			wantResp:    resp.EncodeError("unknown command UNKNOWN"),
			wantCmdName: "UNKNOWN",
			wantStatus:  "err",
		},
		{
			name:        "Wrong Arity",
			setup:       func(db *database.Database) {},
			cmd:         "SET",
			args:        []string{"foo"},
			wantResp:    resp.EncodeError("wrong number of arguments for 'SET'"),
			wantCmdName: "SET",
			wantStatus:  "err",
		},
		{
			name:        "Subcommand Routing",
			setup:       func(db *database.Database) {},
			cmd:         "COMMAND",
			args:        []string{"COUNT"},
			wantCmdName: "COMMAND COUNT",
			wantStatus:  "ok",
		},
		{
			name:        "Unknown Subcommand",
			setup:       func(db *database.Database) {},
			cmd:         "COMMAND",
			args:        []string{"UNKNOWN"},
			wantResp:    resp.EncodeError("unknown subcommand UNKNOWN"),
			wantCmdName: "COMMAND",
			wantStatus:  "err",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := generateSampleDB()
			tt.setup(db)

			got := Dispatch(db, tt.cmd, tt.args)

			if tt.wantResp != nil && !bytes.Equal(got.Response, tt.wantResp) {
				t.Errorf("Dispatch() Response = %q, want %q", got.Response, tt.wantResp)
			}
			if got.CmdName != tt.wantCmdName {
				t.Errorf("Dispatch() CmdName = %v, want %v", got.CmdName, tt.wantCmdName)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Dispatch() Status = %v, want %v", got.Status, tt.wantStatus)
			}
		})
	}
}
