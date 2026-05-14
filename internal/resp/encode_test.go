package resp

import (
	"bytes"
	"testing"
)

func TestEncodeSimpleString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{"OK", "OK", []byte("+OK\r\n")},
		{"PONG", "PONG", []byte("+PONG\r\n")},
		{"Empty", "", []byte("+\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeSimpleString(tt.input); !bytes.Equal(got, tt.want) {
				t.Errorf("EncodeSimpleString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEncodeError(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{"Simple error", "unknown command", []byte("-ERR unknown command\r\n")},
		{"Empty", "", []byte("-ERR \r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeError(tt.input); !bytes.Equal(got, tt.want) {
				t.Errorf("EncodeError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  []byte
	}{
		{"Positive", 42, []byte(":42\r\n")},
		{"Zero", 0, []byte(":0\r\n")},
		{"Negative", -1, []byte(":-1\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeInteger(tt.input); !bytes.Equal(got, tt.want) {
				t.Errorf("EncodeInteger() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEncodeBulkString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{"Normal", "foo", []byte("$3\r\nfoo\r\n")},
		{"Empty", "", []byte("$0\r\n\r\n")},
		{"With spaces", "foo bar", []byte("$7\r\nfoo bar\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBulkString(tt.input); !bytes.Equal(got, tt.want) {
				t.Errorf("EncodeBulkString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEncodeNullBulkString(t *testing.T) {
	want := []byte("$-1\r\n")
	if got := EncodeNullBulkString(); !bytes.Equal(got, want) {
		t.Errorf("EncodeNullBulkString() = %q, want %q", got, want)
	}
}

func TestEncodeArray(t *testing.T) {
	tests := []struct {
		name  string
		input [][]byte
		want  []byte
	}{
		{
			"Two bulk strings",
			[][]byte{EncodeBulkString("foo"), EncodeBulkString("bar")},
			[]byte("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"),
		},
		{
			"Empty array",
			[][]byte{},
			[]byte("*0\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeArray(tt.input); !bytes.Equal(got, tt.want) {
				t.Errorf("EncodeArray() = %q, want %q", got, tt.want)
			}
		})
	}
}
