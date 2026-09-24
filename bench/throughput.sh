#!/usr/bin/env bash
# End-to-end throughput against a running kvgo, over TCP and RESP.
# Numbers include the network path, RESP parsing and, for SET, the WAL.
#
# Usage: REPEATS=5 CLIENTS="1 8 50" ./bench/throughput.sh
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

REQUESTS=${REQUESTS:-1000000}
CLIENTS=${CLIENTS:-"1 2 4 8 16 32 50 64"}
OPS=${OPS:-"get set"}
REPEATS=${REPEATS:-5}
KEYSPACE=${KEYSPACE:-1000000}
KVGO_PORT=${KVGO_PORT:-6379}
SHARD_AMOUNT=${SHARD_AMOUNT:-unknown}

require redis-benchmark
require redis-cli
wait_ready "$KVGO_PORT" 5

OUT=$(out_file throughput)

{
    hardware_info
    echo "shards:  $SHARD_AMOUNT"
    echo "params:  requests=$REQUESTS keyspace=$KEYSPACE repeats=$REPEATS transport=tcp"
    echo "---"
    echo '"op","clients","run","rps","avg_latency_ms","min_latency_ms","p50_latency_ms","p95_latency_ms","p99_latency_ms","max_latency_ms"'

    for op in $OPS; do
        for clients in $CLIENTS; do
            for run in $(seq "$REPEATS"); do
                redis-benchmark -p "$KVGO_PORT" -t "$op" -n "$REQUESTS" -c "$clients" -r "$KEYSPACE" --csv \
                    | grep -i "^\"$op" \
                    | sed "s/^\"[A-Za-z]*\"/\"$op\",\"$clients\",\"$run\"/"
            done
        done
    done
} | tee "$OUT"

echo ""
echo "results saved to $OUT"
