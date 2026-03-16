# kvgo
> A simple, concurrent in-memory KV store

[![Go Version](https://img.shields.io/badge/go-1.25.6-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/rvHoney/kvgo/actions/workflows/ci.yml/badge.svg)](https://github.com/rvHoney/kvgo/actions/workflows/ci.yml)

## Overview
**kvgo** is a lightweight, in-memory key-value store implemented in Go. Built as an educational Redis clone, it focuses on simplicity, clear concurrent patterns, and standard library usage.

## Key Features (Current)
* **TCP Server:** Custom connection handling for concurrent clients.
* **Thread-safe Storage:** Wrapped map with `sync.RWMutex` for safe concurrent operations.
* **Basic Commands:** Support for standard operations (`GET`, `SET`, `DEL`).
* **Sharding:** Partition the storage to reduce lock contention on larger datasets.

## Engineering Focus
The project is built with a focus on straightforward, idiomatic Go patterns:
* **Lock Granularity:** Uses `sync.RWMutex` to safely handle concurrent reads and writes without over-complicating state management.
* **Connection Management:** Spawns a lightweight goroutine per client for simple and effective concurrent request handling.
* **Configuration:** Uses Go's standard `flag` package for configuration without external dependencies.

## Getting Started

Clone the repository and build the binary using standard Makefile directives:

```bash
# Build the binary into ./bin/kvgo-server
make build

# Run the server with configuration flags
make run
```

## Quality Assurance
The project includes automated tests (via a GitHub Actions CI) and checks for race conditions:

```bash
# Run the test suite with the race detector enabled (-race)
make test

# Run tests and generate a HTML coverage report
make coverage
```

## Roadmap
* **RESP Protocol:** Support the REdis Serialization Protocol (RESP) to be compatible with standard Redis clients.
* **Worker Pools:** Transition from 1-goroutine-per-connection to a worker pool to better manage resources under high connection loads.
* **RAFT:** Ensure data replication and integrity thanks to an elected leader.

---
*A project focused on simplicity and Go concurrency.*