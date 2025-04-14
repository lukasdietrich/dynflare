package dyndns

import (
	"fmt"
	"net"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/monitor"
)

type filter struct {
	program *vm.Program
}

func newFilter(cfg config.Domain) (*filter, error) {
	program, err := expr.Compile(cfg.Filter.String(),
		expr.Env(&environment{}),
		expr.AsBool(),
		expr.WarnOnAny(),
	)

	if err != nil {
		return nil, fmt.Errorf("could not compile filter expression of %q:%w", cfg.Name, err)
	}

	return &filter{program}, nil
}

func (f *filter) match(addr monitor.Addr) (bool, error) {
	result, err := expr.Run(f.program, &environment{addr})
	if err != nil {
		return false, fmt.Errorf("could not evaluate filter expression: %w", err)
	}

	resultBool, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("filter expression did not evaluate to a bool: %v", result)
	}

	return resultBool, nil
}

type environment struct {
	monitor.Addr
}

func mustParseIPNet(s string) net.IPNet {
	_, ipnet, _ := net.ParseCIDR(s)
	return *ipnet
}

var specialLocalNetworks = [...]net.IPNet{
	mustParseIPNet("fc00::/7"),  // Unique local address
	mustParseIPNet("fe80::/10"), // Link-local address
}

func (e *environment) IsPublic() bool {
	for _, mask := range specialLocalNetworks {
		if mask.Contains(e.IP) {
			return false
		}
	}

	return e.IP.IsGlobalUnicast()
}

func (e *environment) Is4() bool {
	return len(e.IP) == net.IPv4len
}

func (e *environment) Is6() bool {
	return len(e.IP) == net.IPv6len
}

func (e *environment) IsInterface(iface string) bool {
	return e.LinkName == iface
}

func (e *environment) HasPrefix(prefixStr string) bool {
	var (
		prefix = net.ParseIP(prefixStr)
		mask   = e.Mask
		ip     = e.IP
	)

	for i, maskByte := range mask {
		// mask   = 11111100
		// prefix = 10......
		// ip     = 10110110

		if prefix[i]^ip[i]&maskByte != 0 {
			return false
		}
	}

	return true
}

func (e *environment) HasSuffix(suffixStr string) bool {
	var (
		suffix = net.ParseIP(suffixStr)
		mask   = e.Mask
		ip     = e.IP
	)

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

func (e *environment) HasFlag(flag string) bool {
	return slices.Contains(e.Flags, monitor.Flag(flag))
}
