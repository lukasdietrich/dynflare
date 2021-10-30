package dyndns

import (
	"fmt"
	"log"

	"github.com/lukasdietrich/dynflare/internal/cache"
	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/monitor"
	"github.com/lukasdietrich/dynflare/internal/nameserver"
)

type Updater struct {
	cache       *cache.Cache
	domainSlice []*domainUpdater
}

func NewUpdater(config config.Config, cache *cache.Cache) (*Updater, error) {
	nameserverMap, err := newNameservers(config)
	if err != nil {
		return nil, err
	}

	return &Updater{
		cache:       cache,
		domainSlice: newDomainUpdaters(config, nameserverMap),
	}, nil
}

func (u *Updater) Update(updates <-chan *monitor.State) error {
	for state := range updates {
		log.Print("network configuration changed")

		addrSlice := state.AddrSlice()
		if err := u.updateDomains(addrSlice); err != nil {
			return err
		}
	}

	return nil
}

func (u *Updater) updateDomains(addrSlice []monitor.Addr) error {
	defer u.cache.PersistIfDirty()

	for _, domain := range u.domainSlice {
		if err := domain.update(u.cache, addrSlice); err != nil {
			return err
		}
	}

	return nil
}

func newNameservers(cfg config.Config) (map[string]nameserver.Nameserver, error) {
	nameserverMap := make(map[string]nameserver.Nameserver)
	for _, c := range cfg.Nameservers {
		server, err := nameserver.New(c)
		if err != nil {
			return nil, err
		}

		for _, zone := range c.Zones {
			if _, ok := nameserverMap[zone]; ok {
				return nil, fmt.Errorf("zone %q is defined multiple times", zone)
			}

			nameserverMap[zone] = server
		}
	}

	return nameserverMap, nil
}

func newDomainUpdaters(cfg config.Config, nameserverMap map[string]nameserver.Nameserver) []*domainUpdater {
	domainSlice := make([]*domainUpdater, len(cfg.Domains))
	for i, c := range cfg.Domains {
		domainSlice[i] = &domainUpdater{
			nameserver: nameserverMap[c.Zone],
			filter:     newFilter(c),
			zoneName:   c.Zone,
			domainName: c.Name,
		}
	}

	return domainSlice
}
