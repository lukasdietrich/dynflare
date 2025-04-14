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
			domain: config.Domain{Filter: "Is6()"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24")},
			valid:  false,
		},
		{
			domain: config.Domain{Filter: "Is4()"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24")},
			valid:  true,
		},
		{
			domain: config.Domain{Filter: `Is4() and IsInterface("eth0")`},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24"), LinkName: "eth1"},
			valid:  false,
		},
		{
			domain: config.Domain{Filter: `Is4() and IsInterface("eth0")`},
			addr:   monitor.Addr{IPNet: mustParseIPNet("140.1.2.3/24"), LinkName: "eth0"},
			valid:  true,
		},
		{
			domain: config.Domain{Filter: "Is4()"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("2021:1234:1234:1234:1234:1234:1234:1234/64")},
			valid:  false,
		},
		{
			domain: config.Domain{Filter: "Is6()"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("2021:1234:1234:1234:1234:1234:1234:1234/64")},
			valid:  true,
		},
		{
			domain: config.Domain{Filter: "Is6() and IsPublic()"},
			addr:   monitor.Addr{IPNet: mustParseIPNet("fd00::1234:1234:1234:1234/64")},
			valid:  false,
		},
	} {
		filter, err := newFilter(tc.domain)
		if err != nil {
			t.Errorf("domain=%+v %v", tc.domain, err)
		}

		if valid, err := filter.match(tc.addr); err != nil || valid != tc.valid {
			t.Errorf("domain=%+v addr=%+v expected valid=%v", tc.domain, tc.addr, tc.valid)
		}
	}
}
