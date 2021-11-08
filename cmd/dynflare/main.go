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
		log.Fatal().Err(err).Msg("a fatal error occurred. stopping dynflare.")
	}
}

func run() error {
	var (
		configFilename      string
		cacheFilename       string
		logLevel            string
		enableConsoleLogger bool
	)

	flag.StringVar(&configFilename, "config", "config.toml", "Path to config.toml")
	flag.StringVar(&cacheFilename, "cache", "cache.toml", "Path to cache.toml")
	flag.StringVar(&logLevel, "log", "debug", "Set the log level (debug, info, warn, error)")
	flag.BoolVar(&enableConsoleLogger, "pretty-console-logger", false, "Enable pretty, but inefficient console logger")
	flag.Parse()

	if err := setupLogger(logLevel, enableConsoleLogger); err != nil {
		return fmt.Errorf("could not setup logger: %w", err)
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

func setupLogger(logLevel string, enableConsoleLogger bool) error {
	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		return fmt.Errorf("could not set log level to %q: %w", logLevel, err)
	}

	zerolog.SetGlobalLevel(level)

	if enableConsoleLogger {
		log.Logger = log.Logger.Output(zerolog.NewConsoleWriter())
	}

	return nil
}

func update(config config.Config, cache *cache.Cache) error {
	updater, err := dyndns.NewUpdateManager(config, cache)
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

	updater.HandleUpdates(updates) // <- infinite loop
	return fmt.Errorf("the event loop stopped unexpectedly")
}
