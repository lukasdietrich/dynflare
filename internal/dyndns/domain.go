package dyndns

import (
	"log"
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
}

func (d *domainUpdater) update(cache *cache.Cache, addrSlice []monitor.Addr) error {
	addr := d.filterCandidate(addrSlice)
	if addr != nil && d.checkCache(cache, addr) {
		log.Printf("candidate differs from cache. updating record for %q", d.domainName)

		if err := d.updateRecord(addr); err != nil {
			return err
		}

		d.updateCache(cache, addr)
	}

	return nil
}

func (d *domainUpdater) filterCandidate(addrSlice []monitor.Addr) *monitor.Addr {
	for _, addr := range addrSlice {
		if d.filter.match(addr) {
			log.Printf("found an ip canditate for %q: %q", d.domainName, addr.IP)
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
	parts := []string{
		url.PathEscape(d.domainName),
		strings.ToLower(string(determineIPKind(addr))),
	}

	return strings.Join(parts, ":")
}
