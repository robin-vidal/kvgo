# kvgo

[![Go](https://img.shields.io/badge/go-1.25.6-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/robin-vidal/kvgo/actions/workflows/ci.yml/badge.svg)](https://github.com/robin-vidal/kvgo/actions/workflows/ci.yml)

Sharded in-memory key-value store in Go. Custom TCP server, per-shard `RWMutex` locking, Prometheus/Grafana observability.

## Architecture

```
client (TCP)
  ||
  \/
server.Listen -> goroutine-per-connection
  ||
  \/
command parser -> command router (GET, SET, DEL)
  ||
  \/
ShardedStore (N shards, FNV-32a hash)
  |-> shard[0] { sync.RWMutex, map[string]string }
  |-> shard[1] { ... }
  |-> shard[N-1] { ... }
```

Keys are distributed across `runtime.NumCPU()` shards via FNV-32a so concurrent reads on different shards never contend. Prometheus metrics (latency histograms, hit/miss ratio, active connections, shard distribution) are exposed on `:2112/metrics` with a provisioned Grafana dashboard.

## Commands

| Command | Description |
|---------|-------------|
| `SET key value` | Store a key-value pair |
| `GET key` | Retrieve a value by key |
| `DEL key` | Delete a key |

## Quick start

```bash
make docker-up       # kvgo + Prometheus + Grafana
make run             # local binary only
make test            # go test -v -race ./...
```

- kvgo: `localhost:6379`
- Prometheus: `localhost:9090`
- Grafana: `localhost:3000` (dashboard auto-provisioned)

## Roadmap

- **RESP protocol:** implement the Redis Serialization Protocol for compatibility with standard Redis clients.
- **Worker pool:** bounded goroutine pool to cap resource usage under high connection counts.
- **Raft consensus:** multi-node replication with leader election for fault tolerance.
