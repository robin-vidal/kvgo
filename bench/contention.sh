#!/usr/bin/env bash
# Lock contention against concurrent writers. Every series runs the shipped
# Database.Set, so the only variable is the shard count: 1 shard is the
# single-mutex version, then 8 and 32.
# In-process, so no TCP, no RESP and no WAL: this measures locking, not kvgo's
# end-to-end throughput (see throughput.sh for that).
#
# Usage: COUNT=10 WORKERS="1 2 4 8 16 32 64" ./bench/contention.sh
set -euo pipefail

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

COUNT=${COUNT:-10}
BENCHTIME=${BENCHTIME:-500000x}
SHARDS=${SHARDS:-"1 8 32"}
PKG=${PKG:-./internal/database/}

require go

SHARD_RE=$(echo "$SHARDS" | tr ' ' '|')
OUT=$(out_file contention)
RAW=$(mktemp)
trap 'rm -f "$RAW"' EXIT

cd "$REPO_DIR"

echo "running benchmarks (count=$COUNT, benchtime=$BENCHTIME)..." >&2
go test "$PKG" \
    -run '^$' \
    -bench "BenchmarkDatabaseSet/shards=($SHARD_RE)/workers=" \
    -benchtime "$BENCHTIME" \
    -count "$COUNT" \
    -timeout 60m \
    | tee "$RAW"

{
    hardware_info
    echo "params:  count=$COUNT benchtime=$BENCHTIME shards=$SHARDS transport=in-process"
    echo "---"
    echo '"series","workers","run","ns_per_op","ops_per_sec"'

    awk '
        /^BenchmarkDatabaseSet\/shards=/ {
            split($1, parts, "/")
            sub("shards=", "", parts[2])
            sub("workers=", "", parts[3])
            sub("-[0-9]+$", "", parts[3])
            series = (parts[2] == "1") ? "single mutex" : parts[2] " shards"
            workers = parts[3]
        }
        /ns\/op/ {
            key = series "|" workers
            run[key]++
            printf "\"%s\",\"%s\",\"%d\",\"%s\",\"%d\"\n", series, workers, run[key], $3, (1e9 / $3)
        }
    ' "$RAW"
} > "$OUT"

echo ""
echo "results saved to $OUT"
