package command

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/store"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func ptr(s string) *string { return &s }

func dummyHandler(s *store.Store, args []string) result {
	return result{Response: resp.EncodeSimpleString("OK"), Status: "ok"}
}

func generateSampleDB(t *testing.T) *database.Database {
	t.Helper()
	db, err := database.New(&config.Config{
		Host:        "localhost",
		Port:        6379,
		Debug:       false,
		ShardAmount: 2,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return db
}

func generateSampleStore(t *testing.T, db *database.Database) *store.Store {
	t.Helper()
	w, err := wal.Open(&config.WalConfig{
		WalPath: filepath.Join(t.TempDir(), "wal.log"),
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { w.Close() })
	return store.New(db, w)
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
			db := generateSampleDB(t)
			tt.setup(db)
			st := generateSampleStore(t, db)

			got := Dispatch(st, tt.cmd, tt.args)

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

func TestApply(t *testing.T) {
	tests := []struct {
		name    string
		entry   wal.Entry
		wantErr bool
	}{
		{
			name: "SET",
			entry: wal.Entry{
				SeqNum:  1,
				Command: "SET",
				Key:     "foo",
				Value:   ptr("bar"),
			},
			wantErr: false,
		},
		{
			name: "DEL",
			entry: wal.Entry{
				SeqNum:  2,
				Command: "DEL",
				Key:     "foo",
			},
			wantErr: false,
		},
		{
			name: "SET no value",
			entry: wal.Entry{
				SeqNum:  3,
				Command: "SET",
				Key:     "foo",
				Value:   nil,
			},
			wantErr: true,
		},
		{
			name: "Unknown Command",
			entry: wal.Entry{
				SeqNum:  4,
				Command: "UNKNOWN",
			},
			wantErr: true,
		},
		{
			name:    "Stateless command (no apply)",
			entry:   wal.Entry{SeqNum: 5, Command: "PING"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := generateSampleDB(t)

			got := Apply(db, tt.entry)
			gotError := got != nil

			if gotError != tt.wantErr {
				t.Errorf("Apply() Error = %v, want %v", gotError, tt.wantErr)
			}
		})
	}
}

func TestDispatch_WritesWAL(t *testing.T) {
	db := generateSampleDB(t)
	path := filepath.Join(t.TempDir(), "wal.log")
	w, err := wal.Open(&config.WalConfig{WalPath: path})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer w.Close()
	st := store.New(db, w)

	Dispatch(st, "SET", []string{"foo", "bar"})

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("expected WAL file to contain an entry, got empty file")
	}
}

func TestDispatch_WALFailureDoesNotModifyDB(t *testing.T) {
	db := generateSampleDB(t)
	w, _ := wal.Open(&config.WalConfig{WalPath: filepath.Join(t.TempDir(), "wal.log")})
	w.Close()
	st := store.New(db, w)

	Dispatch(st, "SET", []string{"foo", "bar"})

	if _, ok := db.Get("foo"); ok {
		t.Error("expected db to remain unmodified after WAL append failure")
	}
}
