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

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
)

// executeCommand dispatches the command based on its name and run it.
func executeCommand(db *database.Database, m *metrics, cmd string, args []string) string {
	switch cmd {
	case "SET":
		if len(args) != 2 {
			m.recordCommand("SET", "err")
			return "ERR wrong number of arguments for 'SET'\n"
		}
		db.Set(args[0], args[1])
		m.recordCommand("SET", "ok")
		return "OK\n"
	case "GET":
		if len(args) != 1 {
			m.recordCommand("GET", "err")
			return "ERR wrong number of arguments for 'GET'\n"
		}
		val, ok := db.Get(args[0])
		if !ok {
			m.recordCommand("GET", "ok")
			return "(nil)\n"
		}
		m.recordCommand("GET", "ok")
		return val + "\n"
	case "DEL":
		if len(args) != 1 {
			m.recordCommand("DEL", "err")
			return "ERR wrong number of arguments for 'DEL'\n"
		}
		db.Delete(args[0])
		m.recordCommand("DEL", "ok")
		return "OK\n"
	default:
		m.recordCommand(cmd, "err")
		return fmt.Sprintf("ERR unknown command '%s'\n", cmd)
	}
}

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
func handleConnection(conn net.Conn, db *database.Database, m *metrics) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Debug("failed to close connection", "error", err)
		}
	}()
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

		response := executeCommand(db, m, cmd, args)
		slog.Debug("executed", "cmd", cmd, "args", args, "response", response)

		_, err = fmt.Fprintln(conn, response)
		if err != nil {
			slog.Debug("failed to send response", "error", err)
			break
		}
	}
}

// Start launches a TCP server according to the configuration
func Start(cfg *config.Config, db *database.Database) error {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer func() {
		if err := ln.Close(); err != nil {
			slog.Debug("failed to close listener", "error", err)
		}
	}()

	slog.Info("TCP server is listening", "addr", ln.Addr().String())

	m, err := newMetrics()
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Debug("connection accept failed", "error", err)
			continue
		}

		go handleConnection(conn, db, m)
	}
}
