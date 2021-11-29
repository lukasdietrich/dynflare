package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Log           Log            `toml:"log"`
	Notifications []Notification `toml:"notification"`
	Nameservers   []Nameserver   `toml:"nameserver"`
	Domains       []Domain       `toml:"domain"`
}

func Parse(filename string) (Config, error) {
	cfg := createDefault()

	meta, err := toml.DecodeFile(filename, &cfg)
	if err != nil {
		return cfg, err
	}

	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		err = fmt.Errorf("invalid config keys: %v", undecoded)
	}

	return cfg, err
}

func createDefault() Config {
	return Config{
		Log: Log{
			Level:     "info",
			Format:    "json",
			Caller:    true,
			Timestamp: true,
		},
		Nameservers: nil,
		Domains:     nil,
	}
}
