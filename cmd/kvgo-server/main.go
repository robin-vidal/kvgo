/*
kvgo-server is the main entry point for the KVGo database engine.
It handles CLI flags, initializes the storage, and starts the TCP server.
*/
package main

import (
	"fmt"
	"os"

	"github.com/rvHoney/kvgo/internal/config"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	if cfg.Debug {
		fmt.Printf("Config loaded: %+v\n", cfg)
	} else {
		fmt.Printf("Config loaded\n")
	}
}
