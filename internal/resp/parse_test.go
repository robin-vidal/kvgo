package resp

import (
	"bufio"
	"slices"
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantCommand Command
		wantErr     bool
	}{
		{
			name:  "Complete Command",
			input: "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n",
			wantCommand: Command{
				Name: "GET",
				Args: []string{"foo"},
			},
			wantErr: false,
		},
		{
			name:  "No Args Command",
			input: "*1\r\n$4\r\nPING\r\n",
			wantCommand: Command{
				Name: "PING",
				Args: nil,
			},
			wantErr: false,
		},
		{
			name:        "Empty Array",
			input:       "*0\r\n",
			wantCommand: Command{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			cmd, err := ParseCommand(reader)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCommand() error %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if cmd.Name != tt.wantCommand.Name {
					t.Errorf("ParseCommand() got %v, wantName %v", cmd.Name, tt.wantCommand.Name)
				}
				if !slices.Equal(cmd.Args, tt.wantCommand.Args) {
					t.Errorf("ParseCommand() got %v, wantArgs %v", cmd.Args, tt.wantCommand.Args)
				}
			}
		})
	}
}

func TestParseArray(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantArray []string
		wantErr   bool
	}{
		{
			name:      "Normal Array",
			input:     "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n",
			wantArray: []string{"GET", "foo"},
			wantErr:   false,
		},
		{
			name:      "Wrong Prefix",
			input:     "+2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n",
			wantArray: nil,
			wantErr:   true,
		},
		{
			name:      "Incorrect Size",
			input:     "*5\r\n$3\r\nGET\r\n$3\r\nfoo\r\n",
			wantArray: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			array, err := parseArray(reader)

			if (err != nil) != tt.wantErr {
				t.Fatalf("parseArray() error %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if !slices.Equal(array, tt.wantArray) {
					t.Errorf("parseArray() got %v, wantArray %v", array, tt.wantArray)
				}
			}
		})
	}
}

func TestParseBulkString(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantString string
		wantErr    bool
	}{
		{
			name:       "Normal Bulk String",
			input:      "$3\r\nfoo\r\n",
			wantString: "foo",
			wantErr:    false,
		},
		{
			name:       "Wrong Prefix",
			input:      "+3\r\nfoo\r\n",
			wantString: "",
			wantErr:    true,
		},
		{
			name:       "Incorrect Size",
			input:      "$5\r\nfoo\r\n",
			wantString: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			str, err := parseBulkString(reader)

			if (err != nil) != tt.wantErr {
				t.Fatalf("parseBulkString() error %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if str != tt.wantString {
					t.Errorf("parseBulkString() got %v, wantString %v", str, tt.wantString)
				}
			}
		})
	}
}

func TestParseLen(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:    "Valid Bulk String Prefix",
			input:   "$6\r\n",
			want:    6,
			wantErr: false,
		},
		{
			name:    "Valid Array Prefix",
			input:   "*3\r\n",
			want:    3,
			wantErr: false,
		},
		{
			name:    "Invalid Number",
			input:   "$abc\r\n",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLen(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseLen() error %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseLen() got %v, want %v", got, tt.want)
			}
		})
	}
}
