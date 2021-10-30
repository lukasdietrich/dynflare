package cache

import (
	"log"
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
			log.Print("cache is empty")
			return nil
		}

		return err
	}

	log.Printf("reading cache from %q", c.filename)

	defer f.Close()
	_, err = toml.DecodeReader(f, &c.entries)
	return err
}

func (c *Cache) PersistIfDirty() error {
	if !c.dirty {
		log.Print("cache is not dirty, skip writing")
		return nil
	}

	log.Printf("writing cache to %q", c.filename)

	f, err := os.OpenFile(c.filename, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if err = toml.NewEncoder(f).Encode(c.entries); err == nil {
		c.dirty = false
	}

	return err
}

func (c *Cache) Get(key string) string {
	return c.entries[key]
}

func (c *Cache) Put(key, value string) {
	if c.entries == nil {
		c.entries = make(map[string]string)
	}

	c.dirty = true
	c.entries[key] = value
}
