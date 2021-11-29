package dyndns

import (
	"testing"

	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/monitor"
)

func TestFilter(t *testing.T) {
	type testcase struct {
		domain config.Domain
		addr   monitor.Addr
		valid  bool
	}

	for _, tc := range []testcase{
		{
			domain: config.Domain{Kind: "AAAA"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24")},
			valid:  false,
		},
		{
			domain: config.Domain{Kind: "A"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24")},
			valid:  true,
		},
		{
			domain: config.Domain{Kind: "A", Interface: "eth0"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24"), LinkName: "eth1"},
			valid:  false,
		},
		{
			domain: config.Domain{Kind: "A", Interface: "eth0"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24"), LinkName: "eth0"},
			valid:  true,
		},
		{
			domain: config.Domain{Kind: "A"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("2021:1234:1234:1234:1234:1234:1234:1234/64")},
			valid:  false,
		},
		{
			domain: config.Domain{Kind: "AAAA"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("2021:1234:1234:1234:1234:1234:1234:1234/64")},
			valid:  true,
		},
		{
			domain: config.Domain{Kind: "AAAA"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("fd00::1234:1234:1234:1234/64")},
			valid:  false,
		},
	} {
		if newFilter(tc.domain).match(tc.addr) != tc.valid {
			t.Errorf("domain=%+v addr=%+v expected valid=%v", tc.domain, tc.addr, tc.valid)
		}
	}
}
