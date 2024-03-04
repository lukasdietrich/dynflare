package dyndns

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/lukasdietrich/dynflare/internal/cache"
	"github.com/lukasdietrich/dynflare/internal/monitor"
	"github.com/lukasdietrich/dynflare/internal/nameserver"
)

type domainUpdater struct {
	nameserver nameserver.Nameserver
	filter     *filter
	zoneName   string
	domainName string
	disabled   bool
}

func (d *domainUpdater) update(cache *cache.Cache, notifier *notifier, addrSlice []monitor.Addr) error {
	addr := d.filterCandidate(addrSlice)
	if addr != nil {
		if d.disabled {
			slog.Debug("candidate found, but skipping because of previous errors.",
				slog.String("domain", d.domainName),
				slog.Any("ip", addr.IP))
			return nil
		}

		if d.checkCache(cache, addr) {
			slog.Info("candidate differs from cache. updating record.",
				slog.String("domain", d.domainName),
				slog.Any("ip", addr.IP))

			if err := d.updateRecord(addr); err != nil {
				return err
			}

			go notifier.notify("The dns record of %q has been updated to %q.", d.domainName, addr.IP)
			d.updateCache(cache, addr)
		} else {
			slog.Debug("candidate matches cache. skip update.",
				slog.String("domain", d.domainName),
				slog.Any("ip", addr.IP))
		}
	}

	return nil
}

func (d *domainUpdater) filterCandidate(addrSlice []monitor.Addr) *monitor.Addr {
	for _, addr := range addrSlice {
		if d.filter.match(addr) {
			slog.Debug("found an ip candidate",
				slog.String("domain", d.domainName),
				slog.Any("ip", addr.IP))

			return &addr
		}
	}

	return nil
}

func (d *domainUpdater) updateRecord(addr *monitor.Addr) error {
	record := nameserver.Record{
		Zone:   d.zoneName,
		Domain: d.domainName,
		Kind:   determineIPKind(addr),
		IP:     addr.IP,
	}

	return d.nameserver.UpdateRecord(record)
}

func (d *domainUpdater) checkCache(cache *cache.Cache, addr *monitor.Addr) bool {
	cached := cache.Get(d.deriveCacheKey(addr))
	return cached != addr.IP.String()
}

func (d *domainUpdater) updateCache(cache *cache.Cache, addr *monitor.Addr) {
	cache.Put(d.deriveCacheKey(addr), addr.IP.String())
}

func (d *domainUpdater) deriveCacheKey(addr *monitor.Addr) string {
	return fmt.Sprintf("%s:%s", // "example.com:aaaa"
		url.PathEscape(d.domainName),
		strings.ToLower(string(determineIPKind(addr))),
	)
}
