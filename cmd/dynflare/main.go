package main

import (
	"fmt"
	"log"

	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/dyndns"
	"github.com/lukasdietrich/dynflare/internal/systemd"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("an error occurred: %v", err)
	}
}

func run() error {
	flags := parseFlags()
	cfg, err := config.Parse(flags.ConfigFilename)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	if flags.InstallSystemUnit {
		return systemd.Install(flags.ConfigFilename)
	} else {
		return dyndns.Update(cfg)
	}
}
