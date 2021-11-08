package dyndns

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

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

func (d *domainUpdater) update(cache *cache.Cache, addrSlice []monitor.Addr) error {
	addr := d.filterCandidate(addrSlice)
	if addr != nil {
		if d.disabled {
			log.Debug().
				Str("domain", d.domainName).
				Stringer("ip", addr.IP).
				Msg("candidate found, but skipping because of previous errors.")
			return nil
		}

		if d.checkCache(cache, addr) {
			log.Info().
				Str("domain", d.domainName).
				Stringer("ip", addr.IP).
				Msg("candidate differs from cache. updating record.")

			if err := d.updateRecord(addr); err != nil {
				return err
			}

			d.updateCache(cache, addr)
		} else {
			log.Debug().
				Str("domain", d.domainName).
				Stringer("ip", addr.IP).
				Msg("candidate matches cache. skip update.")
		}
	}

	return nil
}

func (d *domainUpdater) filterCandidate(addrSlice []monitor.Addr) *monitor.Addr {
	for _, addr := range addrSlice {
		if d.filter.match(addr) {
			log.Debug().
				Str("domain", d.domainName).
				Stringer("ip", addr.IP).
				Msg("found an ip candidate")

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
