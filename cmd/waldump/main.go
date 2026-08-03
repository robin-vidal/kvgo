// Command waldump dumps a kvgo snapshot and WAL for durability/compaction debugging.
//
// Usage:
//
//	waldump --walPath=/var/lib/kvgo/wal.log --snapshotPath=/var/lib/kvgo/snapshot.db
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func main() {
	walPath := flag.String("walPath", "", "Path to the WAL log file (omit to skip)")
	snapshotPath := flag.String("snapshotPath", "", "Path to the snapshot file (omit to skip)")
	flag.Parse()

	if *walPath == "" && *snapshotPath == "" {
		fmt.Fprintln(os.Stderr, "nothing to do: pass --walPath and/or --snapshotPath")
		os.Exit(2)
	}

	if *snapshotPath != "" {
		dumpSnapshot(*snapshotPath)
	}
	if *walPath != "" {
		dumpWAL(*walPath)
	}
}

func dumpSnapshot(path string) {
	data, err := wal.ReadSnapshot(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "snapshot:", err)
		os.Exit(1)
	}
	if data == nil {
		fmt.Printf("no snapshot at %s\n", path)
		return
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Printf("%d keys in snapshot %s\n", len(data), path)
	for _, k := range keys {
		fmt.Printf("  key=%q value=%q\n", k, data[k])
	}
}

func dumpWAL(path string) {
	w, err := wal.Open(&config.WalConfig{WalPath: path})
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}

	entries, err := w.Replay()
	if err != nil {
		fmt.Fprintln(os.Stderr, "replay:", err)
		os.Exit(1)
	}

	fmt.Printf("%d entries in WAL %s\n", len(entries), path)
	for _, e := range entries {
		val := "<nil>"
		if e.Value != nil {
			val = *e.Value
		}
		fmt.Printf("  seq=%d %s key=%q value=%q\n", e.SeqNum, e.Command, e.Key, val)
	}
}
