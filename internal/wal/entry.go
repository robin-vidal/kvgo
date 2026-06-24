package wal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
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

type reader struct {
	data []byte
	pos  int
}

func (r *reader) bytes(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, errors.New("entry too short")
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
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

	r := &reader{data: data}

	b, err := r.bytes(seqNumSize)
	if err != nil {
		return Entry{}, err
	}
	seqNum := binary.BigEndian.Uint64(b)

	b, err = r.bytes(commandLenSize)
	if err != nil {
		return Entry{}, err
	}
	commandLen := int(b[0])

	b, err = r.bytes(commandLen)
	if err != nil {
		return Entry{}, err
	}
	command := string(b)

	b, err = r.bytes(keyLenSize)
	if err != nil {
		return Entry{}, err
	}
	keyLen := int(binary.BigEndian.Uint16(b))

	b, err = r.bytes(keyLen)
	if err != nil {
		return Entry{}, err
	}
	key := string(b)

	b, err = r.bytes(1)
	if err != nil {
		return Entry{}, err
	}
	present := int(b[0])

	var value *string
	if present == 1 {
		b, err = r.bytes(valueLenSize)
		if err != nil {
			return Entry{}, err
		}
		valueLen := int(binary.BigEndian.Uint16(b))

		b, err = r.bytes(valueLen)
		if err != nil {
			return Entry{}, err
		}
		s := string(b)
		value = &s
	} else if present != 0 {
		return Entry{}, errors.New("invalid present byte")
	}

	return Entry{SeqNum: seqNum, Command: command, Key: key, Value: value}, nil
}

func decodeOne(r *bufio.Reader) (Entry, int, error) {
	var raw bytes.Buffer
	tr := io.TeeReader(r, &raw)

	readField := func(n int) ([]byte, error) {
		buf := make([]byte, n)
		if _, err := io.ReadFull(tr, buf); err != nil {
			if raw.Len() == 0 {
				return nil, io.EOF
			}

			if err == io.EOF {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
		return buf, nil
	}

	buf, err := readField(seqNumSize)
	if err != nil {
		return Entry{}, 0, err
	}
	seqNum := binary.BigEndian.Uint64(buf)

	buf, err = readField(commandLenSize)
	if err != nil {
		return Entry{}, 0, err
	}
	commandLen := int(buf[0])

	buf, err = readField(commandLen)
	if err != nil {
		return Entry{}, 0, err
	}
	command := string(buf)

	buf, err = readField(keyLenSize)
	if err != nil {
		return Entry{}, 0, err
	}
	keyLen := int(binary.BigEndian.Uint16(buf))

	buf, err = readField(keyLen)
	if err != nil {
		return Entry{}, 0, err
	}
	key := string(buf)

	buf, err = readField(1)
	if err != nil {
		return Entry{}, 0, err
	}
	present := int(buf[0])

	var value *string
	if present == 1 {
		buf, err = readField(valueLenSize)
		if err != nil {
			return Entry{}, 0, err
		}
		valueLen := int(binary.BigEndian.Uint16(buf))

		buf, err = readField(valueLen)
		if err != nil {
			return Entry{}, 0, err
		}
		s := string(buf)
		value = &s
	} else if present != 0 {
		return Entry{}, 0, errors.New("invalid present byte")
	}

	checksumBuf := make([]byte, checksumSize)
	if _, err := io.ReadFull(r, checksumBuf); err != nil {
		if err == io.EOF {
			return Entry{}, 0, io.ErrUnexpectedEOF
		}
		return Entry{}, 0, err
	}
	storedChecksum := binary.BigEndian.Uint32(checksumBuf)

	if crc32.ChecksumIEEE(raw.Bytes()) != storedChecksum {
		return Entry{}, 0, errors.New("checksum mismatch")
	}

	return Entry{
		SeqNum:  seqNum,
		Command: command,
		Key:     key,
		Value:   value,
	}, raw.Len() + checksumSize, nil
}
