package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type DomainKind string

const (
	KindIPv4 = DomainKind("A")
	KindIPv6 = DomainKind("AAAA")
)

type Cloudflare struct {
	Token string `toml:"token"`
}

type Domain struct {
	Zone      string     `toml:"zone"`
	Name      string     `toml:"name"`
	Interface string     `toml:"interface"`
	Kind      DomainKind `toml:"kind"`
	Suffix    string     `toml:"suffix"`
}

type Config struct {
	Cloudflare Cloudflare `toml:"cloudflare"`
	Domains    []Domain   `toml:"domain"`
}

func Parse(filename string) (config Config, err error) {
	meta, err := toml.DecodeFile(filename, &config)
	if err != nil {
		return
	}

	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		err = fmt.Errorf("invalid config keys: %v", undecoded)
	}

	return
}
