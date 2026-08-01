#!/usr/bin/env bash
set -euo pipefail

REQUESTS=${REQUESTS:-1000000}
CLIENTS=${CLIENTS:-50}
KEYSPACE=${KEYSPACE:-1000000}
KVGO_PORT=${KVGO_PORT:-6379}
SHARD_AMOUNT=${SHARD_AMOUNT:-unknown}
RESULTS_DIR=${RESULTS_DIR:-"$(dirname "$0")/results"}

check_deps() {
    if ! command -v redis-benchmark &>/dev/null; then
        echo "error: redis-benchmark not found" >&2
        exit 1
    fi
}

check_server() {
    if ! redis-cli -p "$KVGO_PORT" ping &>/dev/null; then
        echo "error: kvgo not reachable on port $KVGO_PORT" >&2
        exit 1
    fi
}

hardware_info() {
    echo "date:    $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    echo "os:      $(uname -s) $(uname -r)"
    if [[ "$(uname -s)" == "Darwin" ]]; then
        echo "cpu:     $(sysctl -n machdep.cpu.brand_string)"
        echo "cores:   $(sysctl -n hw.logicalcpu)"
        echo "ram:     $(( $(sysctl -n hw.memsize) / 1024 / 1024 / 1024 ))GB"
    else
        echo "cpu:     $(grep -m1 'model name\|Hardware\|CPU part' /proc/cpuinfo | cut -d: -f2 | xargs || echo 'arm64')"
        echo "cores:   $(nproc)"
        echo "ram:     $(( $(grep MemTotal /proc/meminfo | awk '{print $2}') / 1024 / 1024 ))GB"
    fi
    echo "shards:  $SHARD_AMOUNT"
    echo "params:  requests=$REQUESTS clients=$CLIENTS keyspace=$KEYSPACE"
}

mkdir -p "$RESULTS_DIR"

check_deps
check_server

OUT="$RESULTS_DIR/kvgo_throughput.txt"

{
    hardware_info
    echo "---"
    echo '"clients","rps","avg_latency_ms","min_latency_ms","p50_latency_ms","p95_latency_ms","p99_latency_ms","max_latency_ms"'
    redis-benchmark -p "$KVGO_PORT" -t get -n "$REQUESTS" -c "$CLIENTS" -r "$KEYSPACE" --csv \
        | grep '^"GET"' \
        | sed "s/^\"GET\"/\"$CLIENTS\"/"
} | tee "$OUT"

echo ""
echo "results saved to $OUT"
