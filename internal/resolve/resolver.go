package resolve

import (
	"fmt"
	"net"

	"github.com/lukasdietrich/dynflare/internal/config"
)

func Resolve(domain config.Domain) (net.IP, error) {
	ipNetSlice, err := findAllByInterface(domain.Interface)
	if err != nil {
		return nil, err
	}

	matchingIPSlice := filterMatchingIPs(domain, ipNetSlice)
	if l := len(matchingIPSlice); l != 1 {
		debugName := fmt.Sprintf("%q (interface=%q, kind=%q, suffix=%q)",
			domain.Name, domain.Interface, domain.Kind, domain.Suffix)

		if l == 0 {
			return nil, fmt.Errorf("could not find an ip for %s", debugName)
		} else if l > 1 {
			return nil, fmt.Errorf("found more than one ip for %s", debugName)
		}
	}

	return matchingIPSlice[0], nil
}

func findAllByInterface(name string) ([]*net.IPNet, error) {
	networkInterface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("could not find network-interface %q: %w", name, err)
	}

	addrSlice, err := networkInterface.Addrs()
	if err != nil {
		return nil, err
	}

	var ipNetSlice []*net.IPNet
	for _, addr := range addrSlice {
		if ipNet, ok := addr.(*net.IPNet); ok {
			ipNetSlice = append(ipNetSlice, ipNet)
		}
	}

	return ipNetSlice, nil
}
