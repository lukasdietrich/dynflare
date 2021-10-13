package resolve

import (
	"net"

	"github.com/lukasdietrich/dynflare/internal/config"
)

func filterMatchingIPs(domain config.Domain, ipNetSlice []*net.IPNet) []net.IP {
	var suffix net.IP
	if domain.Suffix != "" {
		suffix = net.ParseIP(domain.Suffix)
	}

	var matchingIPSlice []net.IP

	for _, ipNet := range ipNetSlice {
		ip, mask := normalize(ipNet.IP), ipNet.Mask

		if isValidType(domain, ip) && isValidSuffix(suffix, ip, mask) {
			matchingIPSlice = append(matchingIPSlice, ip)
		}
	}

	return matchingIPSlice
}

func normalize(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		return v4
	}

	return ip.To16()
}

func isValidType(domain config.Domain, ip net.IP) bool {
	v4 := domain.Kind == config.KindIPv4 && len(ip) == net.IPv4len
	v6 := domain.Kind == config.KindIPv6 && len(ip) == net.IPv6len

	return v4 || v6
}

func isValidSuffix(suffix net.IP, ip net.IP, mask net.IPMask) bool {
	if suffix == nil {
		return true
	}

	for i, maskByte := range mask {
		// mask   = 11111100
		// suffix = ......10
		// ip     = 10110110

		if suffix[i]^ip[i]&^maskByte != 0 {
			return false
		}
	}

	return true
}
