package server

import "testing"

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
