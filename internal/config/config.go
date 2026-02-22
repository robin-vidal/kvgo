// Package config handles command-line flags parsing.
package config

import (
	"flag"
	"fmt"
	"os"
)

// Config stores kvgo server config such as host, port and debug mode.
type Config struct {
	Host  string
	Port  int
	Debug bool
}

// Parse initializes a Config struct according to startup flags.
func Parse(args []string) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("kvgo", flag.ContinueOnError)

	fs.StringVar(&cfg.Host, "host", "localhost", "The host to bind to")
	fs.IntVar(&cfg.Port, "port", 6379, "The port to listen on")
	fs.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of kvgo:\n")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
