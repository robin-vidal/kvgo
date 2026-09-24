#!/usr/bin/env bash
# Shared helpers for the benchmark scripts.

RESULTS_DIR=${RESULTS_DIR:-"$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/results"}

# stamp returns the UTC datetime used to name result files.
stamp() {
    date -u '+%Y%m%dT%H%M%SZ'
}

# out_file <name> -> results/<datetime>_<name>.txt
out_file() {
    mkdir -p "$RESULTS_DIR"
    echo "$RESULTS_DIR/$(stamp)_$1.txt"
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
    echo "kvgo:    $(git -C "$(dirname "${BASH_SOURCE[0]}")/.." rev-parse --short HEAD 2>/dev/null || echo unknown)"
}

require() {
    if ! command -v "$1" &>/dev/null; then
        echo "error: $1 not found" >&2
        exit 1
    fi
}

# wait_ready <port> [attempts] : blocks until the server answers PING.
wait_ready() {
    local port=$1 attempts=${2:-50}
    for _ in $(seq "$attempts"); do
        if redis-cli -p "$port" ping &>/dev/null; then
            return 0
        fi
        sleep 0.2
    done
    echo "error: kvgo not reachable on port $port" >&2
    return 1
}
