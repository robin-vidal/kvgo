// Package server handles TCP communications
package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"

	"github.com/rvHoney/kvgo/internal/config"
)

// parseCommand parses the user input into a command and its arguments.
func parseCommand(input string) (string, []string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil, errors.New("empty command")
	}

	fields := strings.Fields(input)

	return strings.ToUpper(fields[0]), fields[1:], nil
}

// handleConnection manages a TCP connection, reading and executing commands in a loop.
func handleConnection(conn net.Conn) {
	defer conn.Close()
	slog.Debug("new TCP connection", "remoteAddr", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				slog.Debug("packet reading fail", "error", err)
			}
			break
		}

		cmd, args, err := parseCommand(line)
		if err != nil {
			slog.Debug("packet parsing fail", "error", err)
			continue
		}

		slog.Debug("executing", "cmd", cmd, "args", args)
		fmt.Fprintln(conn, "OK")
	}
}

// Start launches a TCP server according to the configuration
func Start(cfg *config.Config) error {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer ln.Close()

	slog.Info("TCP server is listening", "addr", ln.Addr().String())
	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Debug("connection accept failed", "error", err)
			continue
		}

		go handleConnection(conn)
	}
}
