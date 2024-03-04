package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/lukasdietrich/dynflare/internal/cache"
	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/dyndns"
	"github.com/lukasdietrich/dynflare/internal/monitor"
)

func main() {
	if err := run(); err != nil {
		slog.Error("a fatal error occurred. stopping dynflare.", slog.Any("err", err))
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

	cfg, err := config.Parse(configFilename)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}

	if err := setupLogger(cfg); err != nil {
		return fmt.Errorf("could not setup logger: %w", err)
	}

	cache, err := cache.NewCache(cacheFilename)
	if err != nil {
		return fmt.Errorf("could not open cache: %w", err)
	}

	return update(cfg, cache)
}

func setupLogger(cfg config.Config) error {
	handler, err := createLoggerHandler(cfg)
	if err != nil {
		return fmt.Errorf("could not create logger handler: %w", err)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}

func createLoggerHandler(cfg config.Config) (slog.Handler, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		return nil, fmt.Errorf("could not set log level to %q: %w", cfg.Log.Level, err)
	}

	slog.Info("setting log level", slog.Any("logLevel", level))

	options := slog.HandlerOptions{
		AddSource: cfg.Log.Caller,
		Level:     level,
	}

	if cfg.Log.Format == "json" {
		return slog.NewJSONHandler(os.Stderr, &options), nil
	}

	return slog.NewTextHandler(os.Stderr, &options), nil
}

func update(cfg config.Config, cache *cache.Cache) error {
	updaterManager, err := dyndns.NewUpdateManager(cfg, cache)
	if err != nil {
		return fmt.Errorf("could not create updatemanager: %w", err)
	}

	state, err := monitor.NewState()
	if err != nil {
		return fmt.Errorf("could not create state: %w", err)
	}

	updates, err := state.Monitor()
	if err != nil {
		return fmt.Errorf("could not start state monitor: %w", err)
	}

	updaterManager.HandleUpdates(updates) // <- infinite loop
	return fmt.Errorf("the event loop stopped unexpectedly")
}
