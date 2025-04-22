package dyndns

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"

	"github.com/lukasdietrich/dynflare/internal/cache"
	"github.com/lukasdietrich/dynflare/internal/hook"
	"github.com/lukasdietrich/dynflare/internal/monitor"
	"github.com/lukasdietrich/dynflare/internal/nameserver"
)

type domainUpdater struct {
	nameserver nameserver.Nameserver
	postUp     *hook.Hook
	filter     *filter
	zoneName   string
	domainName string
	comment    string
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

			if err := d.updateRecord(notifier, addr); err != nil {
				return err
			}

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
		normalizeIPNet(&addr.IPNet)

		matches, err := d.filter.match(addr)
		if err != nil {
			slog.Error("could not filter candidate", slog.Any("err", err))
			continue
		}

		if matches {
			slog.Debug("found an ip candidate",
				slog.String("domain", d.domainName),
				slog.Any("ip", addr.IP))

			return &addr
		}
	}

	return nil
}

func (d *domainUpdater) updateRecord(notifier *notifier, addr *monitor.Addr) error {
	record := nameserver.Record{
		Zone:    d.zoneName,
		Domain:  d.domainName,
		Kind:    determineIPKind(addr),
		IP:      addr.IP,
		Comment: d.comment,
	}

	changed, err := d.nameserver.UpdateRecord(record)
	if err != nil {
		return err
	}

	if changed {
		go notifier.notify("The dns record of %q has been updated to %q.", d.domainName, addr.IP)
		go d.callHook(record)
	}

	return nil
}

func (d *domainUpdater) callHook(record nameserver.Record) {
	if err := d.postUp.Execute(record); err != nil {
		slog.Warn("error calling post-up hook", slog.Any("err", err))
	}
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

func normalizeIPNet(ipNet *net.IPNet) {
	if v4 := ipNet.IP.To4(); v4 != nil {
		ipNet.IP = v4
	}
}

func determineIPKind(addr *monitor.Addr) nameserver.RecordKind {
	if len(addr.IP) == net.IPv4len {
		return nameserver.KindV4
	}

	return nameserver.KindV6
}
