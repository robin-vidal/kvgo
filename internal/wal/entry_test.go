package wal

import (
	"slices"
	"strings"
	"testing"
)

func ptr(s string) *string { return &s }

func valueEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
	}{
		{
			name: "simple set",
			entry: Entry{
				SeqNum:  1,
				Command: "SET",
				Key:     "foo",
				Value:   ptr("bar"),
			},
		},
		{
			name: "simple del",
			entry: Entry{
				SeqNum:  2,
				Command: "DEL",
				Key:     "foo",
				Value:   nil,
			},
		},
		{
			name: "empty value set",
			entry: Entry{
				SeqNum:  3,
				Command: "SET",
				Key:     "foo",
				Value:   ptr(""),
			},
		},
		{
			name: "empty key",
			entry: Entry{
				SeqNum:  4,
				Command: "SET",
				Key:     "",
				Value:   ptr("bar"),
			},
		},
		{
			name: "large key and value",
			entry: Entry{
				SeqNum:  5,
				Command: "SET",
				Key:     strings.Repeat("k", 65535),
				Value:   ptr(strings.Repeat("v", 65535)),
			},
		},
		{
			name: "max seq num",
			entry: Entry{
				SeqNum:  uint64(0),
				Command: "SET",
				Key:     "foo",
				Value:   ptr("bar"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.entry)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			decoded, err := Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if decoded.SeqNum != tt.entry.SeqNum {
				t.Errorf("SeqNum = %d, want %d", decoded.SeqNum, tt.entry.SeqNum)
			}
			if decoded.Command != tt.entry.Command {
				t.Errorf("Command = %q, want %q", decoded.Command, tt.entry.Command)
			}
			if decoded.Key != tt.entry.Key {
				t.Errorf("Key = %q, want %q", decoded.Key, tt.entry.Key)
			}
			if !valueEqual(decoded.Value, tt.entry.Value) {
				t.Errorf("Value mismatch")
			}
		})
	}
}

func TestEncodeErrors(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
	}{
		{
			name: "command too long",
			entry: Entry{SeqNum: 1,
				Command: strings.Repeat("x", 256),
				Key:     "foo",
				Value:   ptr("bar"),
			},
		},
		{
			name: "key too long",
			entry: Entry{
				SeqNum:  1,
				Command: "SET",
				Key:     strings.Repeat("k", 65536),
				Value:   ptr("bar"),
			},
		},
		{
			name: "value too long",
			entry: Entry{
				SeqNum:  1,
				Command: "SET",
				Key:     "foo",
				Value:   ptr(strings.Repeat("v", 65536)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Encode(tt.entry); err == nil {
				t.Error("Encode() expected error, got nil")
			}
		})
	}
}

func TestDecodeErrors(t *testing.T) {
	valid, err := Encode(Entry{SeqNum: 1, Command: "SET", Key: "foo", Value: ptr("bar")})
	if err != nil {
		t.Fatalf("setup Encode() error = %v", err)
	}

	tests := []struct {
		name    string
		corrupt func([]byte) []byte
	}{
		{
			name: "corrupted checksum",
			corrupt: func(b []byte) []byte {
				c := slices.Clone(b)
				c[len(c)-1] ^= 0xFF
				return c
			},
		},
		{
			name: "corrupted body",
			corrupt: func(b []byte) []byte {
				c := slices.Clone(b)
				c[0] ^= 0xFF
				return c
			},
		},
		{
			name: "too short",
			corrupt: func(b []byte) []byte {
				return []byte{0x00, 0x01}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Decode(tt.corrupt(valid)); err == nil {
				t.Error("Decode() expected error, got nil")
			}
		})
	}
}
