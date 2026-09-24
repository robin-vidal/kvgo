#!/usr/bin/env bash
# Shard-count sweep over TCP: starts a fresh kvgo per shard count, benchmarks
# it, tears it down.
#
# Usage: SHARDS="1 2 4 8 16 32 64" REPEATS=3 ./bench/sharding.sh
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SHARDS=${SHARDS:-"1 2 4 8 16 32 64 128 256"}
CLIENTS=${CLIENTS:-50}
REQUESTS=${REQUESTS:-500000}
OPS=${OPS:-"get set"}
REPEATS=${REPEATS:-3}
KEYSPACE=${KEYSPACE:-1000000}
KVGO_PORT=${KVGO_PORT:-7379}

require redis-benchmark
require redis-cli
require go

WORK_DIR=$(mktemp -d)
BIN="$WORK_DIR/kvgo-server"
SERVER_PID=""

cleanup() {
    if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
    rm -rf "$WORK_DIR"
}
trap cleanup EXIT

echo "building kvgo-server..." >&2
go build -o "$BIN" "$REPO_DIR/cmd/kvgo-server"

start_server() {
    local shards=$1
    local data_dir="$WORK_DIR/data-$shards"
    mkdir -p "$data_dir"

    "$BIN" \
        -port "$KVGO_PORT" \
        -shardAmount "$shards" \
        -nodeID "bench" \
        -walPath "$data_dir/wal.log" \
        -snapshotPath "$data_dir/snapshot.db" \
        -raftStatePath "$data_dir/raft.state" \
        &>"$data_dir/server.log" &

    SERVER_PID=$!
    if ! wait_ready "$KVGO_PORT"; then
        echo "--- server log ---" >&2
        cat "$data_dir/server.log" >&2
        exit 1
    fi
}

stop_server() {
    if [[ -n "$SERVER_PID" ]]; then
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
        SERVER_PID=""
    fi
}

OUT=$(out_file sharding)

{
    hardware_info
    echo "params:  requests=$REQUESTS clients=$CLIENTS keyspace=$KEYSPACE repeats=$REPEATS transport=tcp"
    echo "---"
    echo '"op","shards","run","rps","avg_latency_ms","min_latency_ms","p50_latency_ms","p95_latency_ms","p99_latency_ms","max_latency_ms"'

    for shards in $SHARDS; do
        start_server "$shards"
        for op in $OPS; do
            for run in $(seq "$REPEATS"); do
                redis-benchmark -p "$KVGO_PORT" -t "$op" -n "$REQUESTS" -c "$CLIENTS" -r "$KEYSPACE" --csv \
                    | grep -i "^\"$op" \
                    | sed "s/^\"[A-Za-z]*\"/\"$op\",\"$shards\",\"$run\"/"
            done
        done
        stop_server
    done
} | tee "$OUT"

echo ""
echo "results saved to $OUT"
