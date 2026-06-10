package wal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const (
	commandMaxLen  int = 255
	keyMaxLen      int = 65535
	valueMaxLen    int = 65535
	seqNumSize     int = 8 // uint64
	commandLenSize int = 1 // uint8
	keyLenSize     int = 2
	valueLenSize   int = 2
	checksumSize   int = 4 // uint32
)

type Entry struct {
	SeqNum  uint64
	Command string
	Key     string
	Value   *string
}

func Encode(e Entry) ([]byte, error) {
	if len(e.Command) > commandMaxLen {
		return nil, errors.New("command too long")
	}
	if len(e.Key) > keyMaxLen {
		return nil, errors.New("key too long")
	}
	if e.Value != nil && len(*e.Value) > valueMaxLen {
		return nil, errors.New("value too long")
	}

	var buf bytes.Buffer

	binary.Write(&buf, binary.BigEndian, e.SeqNum)

	buf.WriteByte(byte(len(e.Command)))
	buf.WriteString(e.Command)

	binary.Write(&buf, binary.BigEndian, uint16(len(e.Key)))
	buf.WriteString(e.Key)

	if e.Value == nil {
		buf.WriteByte(0)
	} else {
		buf.WriteByte(1)
		binary.Write(&buf, binary.BigEndian, uint16(len(*e.Value)))
		buf.WriteString(*e.Value)
	}

	checksum := crc32.ChecksumIEEE(buf.Bytes())
	binary.Write(&buf, binary.BigEndian, checksum)

	return buf.Bytes(), nil
}

func Decode(b []byte) (Entry, error) {
	if len(b) < checksumSize {
		return Entry{}, errors.New("entry too short")
	}

	data := b[:len(b)-checksumSize]
	storedChecksum := binary.BigEndian.Uint32(b[len(data):])

	if crc32.ChecksumIEEE(data) != storedChecksum {
		return Entry{}, errors.New("checksum mismatch")
	}

	offset := 0

	seqNum := binary.BigEndian.Uint64(data[offset:])
	offset += seqNumSize

	commandLen := int(data[offset])
	offset += commandLenSize
	command := string(data[offset : offset+commandLen])
	offset += commandLen

	keyLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += keyLenSize
	key := string(data[offset : offset+keyLen])
	offset += keyLen

	present := data[offset]
	offset++

	var value *string
	if present == 1 {
		valueLen := int(binary.BigEndian.Uint16(data[offset:]))
		offset += valueLenSize
		s := string(data[offset : offset+valueLen])
		value = &s
		offset += valueLen
	}

	return Entry{SeqNum: seqNum, Command: command, Key: key, Value: value}, nil
}
