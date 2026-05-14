// Package server handles TCP communications
package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/resp"
)

// executeCommand dispatches the command based on its name and run it.
func executeCommand(db *database.Database, m *metrics, cmd resp.Command) []byte {
	switch cmd.Name {
	case "SET":
		if len(cmd.Args) != 2 {
			m.recordCommand("SET", "err")
			return resp.EncodeError("wrong number of arguments for 'SET'")
		}

		db.Set(cmd.Args[0], cmd.Args[1])
		m.recordCommand("SET", "ok")
		return resp.EncodeSimpleString("OK")
	case "GET":
		if len(cmd.Args) != 1 {
			m.recordCommand("GET", "err")
			return resp.EncodeError("wrong number of arguments for 'GET'")
		}

		val, ok := db.Get(cmd.Args[0])
		if !ok {
			m.recordCommand("GET", "miss")
			return resp.EncodeNullBulkString()
		}

		m.recordCommand("GET", "ok")
		return resp.EncodeBulkString(val)
	case "DEL":
		if len(cmd.Args) != 1 {
			m.recordCommand("DEL", "err")
			return resp.EncodeError("wrong number of arguments for 'DEL'")
		}

		db.Delete(cmd.Args[0])
		m.recordCommand("DEL", "ok")
		return resp.EncodeInteger(1)
	case "PING":
		m.recordCommand("PING", "ok")
		return resp.EncodeSimpleString("PONG")
	case "COMMAND":
		// TODO: return COMMAND DOCS
		if len(cmd.Args) == 0 {
			m.recordCommand("COMMAND", "ok")
			return resp.EncodeArray([][]byte{})
		}

		switch cmd.Args[0] {
		case "DOCS":
			// TODO: return COMMAND DOCS
			m.recordCommand("COMMAND DOCS", "ok")
			return resp.EncodeArray([][]byte{})
		case "COUNT":
			m.recordCommand("COMMAND COUNT", "ok")
			return resp.EncodeInteger(5)
		default:
			m.recordCommand(cmd.Name, "err")
			return resp.EncodeError("unknown subcommand '" + cmd.Args[0] + "'.")

		}
	default:
		m.recordCommand(cmd.Name, "err")
		return resp.EncodeError("unknown command " + string(cmd.Name))
	}
}

// handleConnection manages a TCP connection, reading and executing commands in a loop.
func handleConnection(conn net.Conn, db *database.Database, m *metrics) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Debug("failed to close connection", "error", err)
		}
		m.recordConnection(-1)
	}()
	slog.Debug("new TCP connection", "remoteAddr", conn.RemoteAddr())
	m.recordConnection(1)

	reader := bufio.NewReader(conn)

	for {
		cmd, err := resp.ParseCommand(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Debug("client disconnected", "remoteAddr", conn.RemoteAddr())
			} else {
				slog.Debug("packet parsing fail", "error", err)
			}
			break
		}

		start := time.Now()
		response := executeCommand(db, m, cmd)
		m.recordDuration(cmd.Name, float64(time.Since(start).Microseconds()))

		slog.Debug("executed", "cmd", cmd, "response", response)

		_, err = conn.Write(response)
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

	m, err := newMetrics(db)
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
