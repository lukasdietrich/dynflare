package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/lukasdietrich/dynflare/internal/cache"
	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/dyndns"
	"github.com/lukasdietrich/dynflare/internal/monitor"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("an error occurred: %v", err)
	}
}

func run() error {
	var (
		configFilename string
		cacheFilename  string
	)

	flag.StringVar(&configFilename, "config", "config.toml", "Path to config.toml")
	flag.StringVar(&cacheFilename, "cache", "cache.toml", "Path to cache.toml")
	flag.Parse()

	config, err := config.Parse(configFilename)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	cache, err := cache.NewCache(cacheFilename)
	if err != nil {
		return fmt.Errorf("could not open cache: %w", err)
	}

	return update(config, cache)
}

func update(config config.Config, cache *cache.Cache) error {
	updater, err := dyndns.NewUpdater(config, cache)
	if err != nil {
		return fmt.Errorf("could not create updater: %w", err)
	}

	state, err := monitor.NewState()
	if err != nil {
		return fmt.Errorf("could not create state: %w", err)
	}

	updates, err := state.Monitor()
	if err != nil {
		return fmt.Errorf("could not start state monitor: %w", err)
	}

	return updater.Update(updates)
}
