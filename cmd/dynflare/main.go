package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lukasdietrich/dynflare/internal/cache"
	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/dyndns"
	"github.com/lukasdietrich/dynflare/internal/monitor"
)

func init() {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
}

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("an fatal error occurred. stopping dynflare.")
	}
}

func run() error {
	var (
		configFilename string
		cacheFilename  string
		logLevel       string
	)

	flag.StringVar(&configFilename, "config", "config.toml", "Path to config.toml")
	flag.StringVar(&cacheFilename, "cache", "cache.toml", "Path to cache.toml")
	flag.StringVar(&logLevel, "log", "debug", "Set the log level (debug, info, warn, error)")
	flag.Parse()

	if err := setLogLevel(logLevel); err != nil {
		return fmt.Errorf("could not set log level to %q: %w", logLevel, err)
	}

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

func setLogLevel(logLevel string) error {
	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(level)
	return nil
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
