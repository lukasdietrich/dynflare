package monitor

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

func subscribeNetlink(linkUpdates chan<- netlink.LinkUpdate, addrUpdates chan<- netlink.AddrUpdate) error {
	if err := netlink.LinkSubscribe(linkUpdates, nil); err != nil {
		return fmt.Errorf("could not subscribe to link updates via netlink: %w", err)
	}

	if err := netlink.AddrSubscribe(addrUpdates, nil); err != nil {
		return fmt.Errorf("could not subscribe to address updates via netlink: %w", err)
	}

	return nil
}

func readInitialLinks(s *State) error {
	linkSlice, err := netlink.LinkList()
	if err != nil {
		return err
	}

	for _, link := range linkSlice {
		s.updateLink(netlink.LinkUpdate{Link: link})

		if err := readInitialAddrs(s, link); err != nil {
			return err
		}
	}

	return nil
}

func readInitialAddrs(s *State, link netlink.Link) error {
	addrSlice, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return err
	}

	for _, addr := range addrSlice {
		s.updateAddr(netlink.AddrUpdate{
			LinkAddress: *addr.IPNet,
			LinkIndex:   link.Attrs().Index,
			Flags:       addr.Flags,
			Scope:       addr.Scope,
			PreferedLft: addr.PreferedLft,
			ValidLft:    addr.ValidLft,
			NewAddr:     true,
		})
	}

	return nil
}
