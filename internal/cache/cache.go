package cache

import (
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

type Cache struct {
	filename string
	dirty    bool
	entries  map[string]string
}

func NewCache(filename string) (*Cache, error) {
	cache := &Cache{filename: filename}
	return cache, cache.load()
}

func (c *Cache) load() error {
	f, err := os.Open(c.filename)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("cache is empty")
			return nil
		}

		return err
	}

	slog.Debug("reading cache", slog.String("filename", c.filename))

	defer f.Close()
	_, err = toml.NewDecoder(f).Decode(&c.entries)
	return err

}

func (c *Cache) PersistIfDirty() error {
	if !c.dirty {
		slog.Debug("cache is not dirty, skip writing")
		return nil
	}

	slog.Info("writing cache", slog.String("filename", c.filename))

	f, err := os.OpenFile(c.filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		slog.Error("could not write cache", slog.Any("err", err))
		return err
	}

	defer f.Close()

	if err = toml.NewEncoder(f).Encode(c.entries); err != nil {
		slog.Error("could not encode cache entries", slog.Any("err", err))
	} else {
		slog.Debug("clearing cache dirty flag")
		c.dirty = false
	}

	return err
}

func (c *Cache) Get(key string) string {
	value := c.entries[key]

	slog.Debug("read value from cache",
		slog.String("key", key),
		slog.String("value", value))

	return value
}

func (c *Cache) Put(key, value string) {
	if c.entries == nil {
		slog.Debug("cache is nil, creating new map")
		c.entries = make(map[string]string)
	}

	slog.Debug("updating cache in memory",
		slog.String("key", key),
		slog.String("value", value))

	c.entries[key] = value

	if !c.dirty {
		slog.Debug("setting cache dirty flag")
		c.dirty = true
	}
}
