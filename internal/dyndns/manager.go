package dyndns

import (
	"fmt"
	"log/slog"

	"github.com/lukasdietrich/dynflare/internal/cache"
	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/monitor"
	"github.com/lukasdietrich/dynflare/internal/nameserver"
)

type UpdateManager struct {
	cache        *cache.Cache
	updaterSlice []*domainUpdater
	notifier     *notifier
}

func NewUpdateManager(cfg config.Config, cache *cache.Cache) (*UpdateManager, error) {
	updaterSlice, err := createDomainUpdaters(cfg)
	if err != nil {
		return nil, err
	}

	notifier, err := newNotifier(cfg)
	if err != nil {
		return nil, err
	}

	return &UpdateManager{cache, updaterSlice, notifier}, nil
}

func (u *UpdateManager) HandleUpdates(updates <-chan *monitor.State) {
	for state := range updates {
		slog.Debug("network configuration changed")
		u.updateDomains(state.AddrSlice())
	}
}

func (u *UpdateManager) updateDomains(addrSlice []monitor.Addr) {
	defer u.cache.PersistIfDirty()

	for _, updater := range u.updaterSlice {
		if err := updater.update(u.cache, u.notifier, addrSlice); err != nil {
			slog.Error("could not update domain",
				slog.Any("err", err),
				slog.String("domain", updater.domainName))

			if nameserver.IsPermanentClientError(err) {
				slog.Info("update failed with a permanent client error. disabling updater for this domain to prevent flooding",
					slog.String("domain", updater.domainName))

				updater.disabled = true
			}
		}
	}
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

func createDomainUpdaters(cfg config.Config) ([]*domainUpdater, error) {
	nameserverMap, err := newNameservers(cfg)
	if err != nil {
		return nil, err
	}

	domainSlice := make([]*domainUpdater, len(cfg.Domains))

	for i, c := range cfg.Domains {
		domainSlice[i] = &domainUpdater{
			nameserver: nameserverMap[c.Zone],
			filter:     newFilter(c),
			zoneName:   c.Zone,
			domainName: c.Name,
		}
	}

	return domainSlice, nil
}
