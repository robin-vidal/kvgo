package resp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

// ParseCommand reads a RESP2 command from the reader.
// It expects an array of bulk strings.
func ParseCommand(reader *bufio.Reader) (Command, error) {
	args, err := parseArray(reader)
	if err != nil {
		return Command{}, err
	}

	if len(args) == 0 {
		return Command{}, errors.New("empty command")
	}

	return Command{Name: args[0], Args: args[1:]}, nil
}

func parseArray(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if line[0] != '*' {
		return nil, errors.New("expected *")
	}

	size, err := parseLen(line)
	if err != nil {
		return nil, err
	}

	args := make([]string, size)
	for i := range size {
		args[i], err = parseBulkString(reader)
		if err != nil {
			return nil, err
		}
	}

	return args, nil
}

func parseBulkString(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	if line[0] != '$' {
		return "", errors.New("expected $")
	}

	size, err := parseLen(line)
	if err != nil {
		return "", err
	}

	buf := make([]byte, size)
	if _, err = io.ReadFull(reader, buf); err != nil {
		return "", err
	}

	if _, err = reader.ReadString('\n'); err != nil {
		return "", err
	}

	return string(buf), nil
}

func parseLen(line string) (int, error) {
	return strconv.Atoi(strings.TrimRight(line[1:], "\r\n"))
}
