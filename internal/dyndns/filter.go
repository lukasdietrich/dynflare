package dyndns

import (
	"net"

	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/monitor"
	"github.com/lukasdietrich/dynflare/internal/nameserver"
)

type filter struct {
	kind          nameserver.RecordKind
	interfaceName string
	suffix        net.IP
}

func newFilter(cfg config.Domain) *filter {
	return &filter{
		kind:          nameserver.RecordKind(cfg.Kind),
		interfaceName: cfg.Interface.String(),
		suffix:        net.ParseIP(cfg.Suffix.String()), // net.ParseIP already handles empty string
	}
}

func (f *filter) match(addr monitor.Addr) bool {
	normalizeIPNet(&addr.IPNet)

	return isPublicIP(addr.IP) &&
		f.matchKind(addr) &&
		f.matchInterface(addr) &&
		f.matchSuffix(addr)
}

func mustParseIPNet(s string) net.IPNet {
	_, ipnet, _ := net.ParseCIDR(s)
	return *ipnet
}

var specialLocalNetworks = [...]net.IPNet{
	mustParseIPNet("fc00::/7"),  // Unique local address
	mustParseIPNet("fe80::/10"), // Link-local address
}

func isPublicIP(ip net.IP) bool {
	for _, mask := range specialLocalNetworks {
		if mask.Contains(ip) {
			return false
		}
	}

	return ip.IsGlobalUnicast()
}

func (f *filter) matchKind(addr monitor.Addr) bool {
	return f.kind == determineIPKind(&addr)
}

func (f *filter) matchInterface(addr monitor.Addr) bool {
	return f.interfaceName == "" || f.interfaceName == addr.LinkName
}

func (f *filter) matchSuffix(addr monitor.Addr) bool {
	if f.suffix != nil {
		var (
			suffix = f.suffix
			mask   = addr.Mask
			ip     = addr.IP
		)

		for i, maskByte := range mask {
			// mask   = 11111100
			// suffix = ......10
			// ip     = 10110110

			if suffix[i]^ip[i]&^maskByte != 0 {
				return false
			}
		}
	}

	return true
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
