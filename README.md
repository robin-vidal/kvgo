# kvgo

[![Go](https://img.shields.io/badge/go-1.26.4-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/robin-vidal/kvgo/actions/workflows/ci.yml/badge.svg)](https://github.com/robin-vidal/kvgo/actions/workflows/ci.yml)

Sharded in-memory key-value store in Go. Custom TCP server, per-shard `RWMutex` locking, RESP2 protocol, write-ahead log for durability, Prometheus/Grafana observability.

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
WAL (append-only, fsync, crash recovery, periodic snapshot + compaction)
  ||
  \/
ShardedStore (N shards, FNV-64a hash)
  |-> shard[0] { sync.RWMutex, map[string]string }
  |-> shard[1] { ... }
  |-> shard[N-1] { ... }
```

Keys are distributed across `runtime.NumCPU()` shards via FNV-64a so concurrent reads on different shards never contend. Prometheus metrics (latency histograms, hit/miss ratio, active connections, shard distribution) are exposed on `:2112/metrics` with a provisioned Grafana dashboard.

## Durability (WAL)

Every `SET`/`DEL` is appended to a write-ahead log before being applied to the in-memory store, with `fsync` after each write.

On startup, kvgo loads the latest snapshot (if any) and replays the WAL to rebuild state. Corrupted trailing entries from a crash mid-write are detected via CRC32 and safely truncated rather than blocking startup.

Periodically (configurable threshold), the in-memory state is snapshotted to disk (`encoding/gob`) and the WAL is compacted, keeping recovery fast even after millions of operations.

## Tools

`waldump` inspects on-disk persistence for durability and compaction debugging. Pass either or both paths; each is skipped when omitted.

```bash
go run ./cmd/waldump --snapshotPath=/var/lib/kvgo/snapshot.db --walPath=/var/lib/kvgo/wal.log
```

## Commands

| Command | Description |
|---------|-------------|
| `SET key value` | Store a key-value pair |
| `GET key` | Retrieve a value by key |
| `DEL key` | Delete a key |
| `PING` | Test server liveness |
| `COMMAND COUNT` | Return the number of supported commands |
| `COMMAND LIST` | List all supported commands |

## Quick start

```bash
make docker-up       # kvgo + Prometheus + Grafana
make run             # local binary only
make test            # go test -v -race ./...
```

```bash
redis-cli -p 6379 SET foo bar
redis-cli -p 6379 GET foo
```

- kvgo: `localhost:6379`
- Prometheus: `localhost:9090`
- Grafana: `localhost:3000` (dashboard auto-provisioned)

## Roadmap

- **Raft consensus:** multi-node replication with leader election and log replication for fault tolerance, building on the WAL above.
- **Worker pool:** bounded goroutine pool to cap resource usage under high connection counts.
