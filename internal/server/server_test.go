package server

import (
	"testing"

	"github.com/rvHoney/kvgo/internal/config"
	"github.com/rvHoney/kvgo/internal/database"
)

func generateSampleConfig() *config.Config {
	return &config.Config{
		Host:        "localhost",
		Port:        6379,
		Debug:       false,
		ShardAmount: 2,
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantCmd string
		wantLen int
		wantErr bool
	}{
		{
			name:    "Simple SET",
			input:   "SET key val\n",
			wantCmd: "SET",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "Lowercase to Uppercase",
			input:   "get mykey\r\n",
			wantCmd: "GET",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "Extra spaces",
			input:   "  DEL    key1   ",
			wantCmd: "DEL",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "Empty",
			input:   "\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, err := parseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cmd != tt.wantCmd {
					t.Errorf("got cmd %v, want %v", cmd, tt.wantCmd)
				}
				if len(args) != tt.wantLen {
					t.Errorf("got args len %v, want %v", len(args), tt.wantLen)
				}
			}
		})
	}
}

func TestExecuteCommand(t *testing.T) {
	cfg := generateSampleConfig()
	db := database.New(cfg)

	tests := []struct {
		name string
		cmd  string
		args []string
		want string
	}{
		{
			name: "SET successful",
			cmd:  "SET",
			args: []string{"key1", "value1"},
			want: "OK\n",
		},
		{
			name: "GET successful",
			cmd:  "GET",
			args: []string{"key1"},
			want: "value1\n",
		},
		{
			name: "GET non-existent",
			cmd:  "GET",
			args: []string{"key2"},
			want: "(nil)\n",
		},
		{
			name: "DEL successful",
			cmd:  "DEL",
			args: []string{"key1"},
			want: "OK\n",
		},
		{
			name: "SET wrong args",
			cmd:  "SET",
			args: []string{"key1"},
			want: "ERR wrong number of arguments for 'SET'\n",
		},
		{
			name: "GET wrong args",
			cmd:  "GET",
			args: []string{},
			want: "ERR wrong number of arguments for 'GET'\n",
		},
		{
			name: "DEL wrong args",
			cmd:  "DEL",
			args: []string{},
			want: "ERR wrong number of arguments for 'DEL'\n",
		},
		{
			name: "Unknown command",
			cmd:  "UNKNOWN",
			args: []string{},
			want: "ERR unknown command 'UNKNOWN'\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executeCommand(db, tt.cmd, tt.args)
			if got != tt.want {
				t.Errorf("executeCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
