package cache

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
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
			log.Debug().Msg("cache is empty")
			return nil
		}

		return err
	}

	log.Debug().Str("filename", c.filename).Msg("reading cache")

	defer f.Close()
	_, err = toml.DecodeReader(f, &c.entries)
	return err
}

func (c *Cache) PersistIfDirty() error {
	if !c.dirty {
		log.Debug().Msg("cache is not dirty, skip writing")
		return nil
	}

	log.Info().Str("filename", c.filename).Msg("writing cache")

	f, err := os.OpenFile(c.filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Error().Err(err).Msg("could not write cache")
		return err
	}

	defer f.Close()

	if err = toml.NewEncoder(f).Encode(c.entries); err != nil {
		log.Error().Err(err).Msg("could not encode cache entries")
	} else {
		log.Debug().Msg("clearing cache dirty flag")
		c.dirty = false
	}

	return err
}

func (c *Cache) Get(key string) string {
	value := c.entries[key]
	log.Debug().Str("key", key).Str("value", value).Msg("read value from cache")
	return value
}

func (c *Cache) Put(key, value string) {
	if c.entries == nil {
		log.Debug().Msg("cache is nil, creating new map")
		c.entries = make(map[string]string)
	}

	log.Debug().Str("key", key).Str("value", value).Msg("updating cache in memory")
	c.entries[key] = value

	if !c.dirty {
		log.Debug().Msg("setting cache dirty flag")
		c.dirty = true
	}
}
