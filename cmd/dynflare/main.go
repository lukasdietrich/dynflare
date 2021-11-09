package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	ctx := zerolog.New(createLogWriter(cfg)).With()

	if cfg.Log.Timestamp {
		ctx = ctx.Timestamp()
	}

	if cfg.Log.Caller {
		ctx = ctx.Caller()
	}

	logger := ctx.Logger()

	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Log.Level))
	if err != nil {
		return fmt.Errorf("could not set log level to %q: %w", cfg.Log.Level, err)
	}

	log.Logger = logger.Level(level)

	log.Info().
		Stringer("loglevel", level).
		Msg("setting log level")
	return nil
}

func createLogWriter(cfg config.Config) io.Writer {
	if strings.ToLower(cfg.Log.Format) == "text" {
		w := zerolog.NewConsoleWriter()
		w.TimeFormat = time.RFC3339

		return w
	}

	return os.Stderr
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
