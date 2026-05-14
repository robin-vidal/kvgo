package server

import (
	"bytes"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
)

func generateSampleConfig() *config.Config {
	return &config.Config{
		Host:        "localhost",
		Port:        6379,
		Debug:       false,
		ShardAmount: 2,
	}
}

func TestExecuteCommand(t *testing.T) {
	cfg := generateSampleConfig()
	db := database.New(cfg)

	tests := []struct {
		name string
		cmd  resp.Command
		want []byte
	}{
		{
			name: "SET successful",
			cmd:  resp.Command{Name: "SET", Args: []string{"key1", "value1"}},
			want: []byte("+OK\r\n"),
		},
		{
			name: "GET successful",
			cmd:  resp.Command{Name: "GET", Args: []string{"key1"}},
			want: []byte("$6\r\nvalue1\r\n"),
		},
		{
			name: "GET non-existent",
			cmd:  resp.Command{Name: "GET", Args: []string{"key2"}},
			want: []byte("$-1\r\n"),
		},
		{
			name: "DEL successful",
			cmd:  resp.Command{Name: "DEL", Args: []string{"key1"}},
			want: []byte(":1\r\n"),
		},
		{
			name: "SET wrong args",
			cmd:  resp.Command{Name: "SET", Args: []string{"key1"}},
			want: []byte("-ERR wrong number of arguments for 'SET'\r\n"),
		},
		{
			name: "GET wrong args",
			cmd:  resp.Command{Name: "GET", Args: []string{}},
			want: []byte("-ERR wrong number of arguments for 'GET'\r\n"),
		},
		{
			name: "DEL wrong args",
			cmd:  resp.Command{Name: "DEL", Args: []string{}},
			want: []byte("-ERR wrong number of arguments for 'DEL'\r\n"),
		},
		{
			name: "Unknown command",
			cmd:  resp.Command{Name: "UNKNOWN", Args: []string{}},
			want: []byte("-ERR unknown command UNKNOWN\r\n"),
		},
		{
			name: "PING",
			cmd:  resp.Command{Name: "PING", Args: []string{}},
			want: []byte("+PONG\r\n"),
		},
		{
			name: "COMMAND no args",
			cmd:  resp.Command{Name: "COMMAND", Args: []string{}},
			want: []byte("*0\r\n"),
		},
		{
			name: "COMMAND DOCS",
			cmd:  resp.Command{Name: "COMMAND", Args: []string{"DOCS"}},
			want: []byte("*0\r\n"),
		},
		{
			name: "COMMAND COUNT",
			cmd:  resp.Command{Name: "COMMAND", Args: []string{"COUNT"}},
			want: []byte(":5\r\n"),
		},
		{
			name: "COMMAND unknown subcommand",
			cmd:  resp.Command{Name: "COMMAND", Args: []string{"FOO"}},
			want: []byte("-ERR unknown subcommand 'FOO'.\r\n"),
		},
	}

	m, err := newMetrics(db)
	if err != nil {
		t.Fatalf("newMetrics() error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executeCommand(db, m, tt.cmd)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("executeCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
