package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Nameservers []Nameserver `toml:"nameserver"`
	Domains     []Domain     `toml:"domain"`
}

func Parse(filename string) (config Config, err error) {
	log.Debug().Str("filename", filename).Msg("parsing config")

	meta, err := toml.DecodeFile(filename, &config)
	if err != nil {
		return
	}

	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		err = fmt.Errorf("invalid config keys: %v", undecoded)
	}

	return
}
