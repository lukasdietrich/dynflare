package monitor

import (
	"log/slog"
	"net"

	"github.com/vishvananda/netlink"
)

type Addr struct {
	net.IPNet
	LinkName string
}

type linkState struct {
	Index int
	Name  string
}

type addrState struct {
	net.IPNet
	LinkIndex int
}

type State struct {
	linkMap map[int]linkState    // index -> link
	addrMap map[string]addrState // ip -> addr
}

func NewState() (*State, error) {
	s := State{
		linkMap: make(map[int]linkState),
		addrMap: make(map[string]addrState),
	}

	return &s, readInitialLinks(&s)
}

func (s *State) AddrSlice() []Addr {
	addrSlice := make([]Addr, 0, len(s.addrMap))

	for _, addr := range s.addrMap {
		addrSlice = append(addrSlice, Addr{
			IPNet:    addr.IPNet,
			LinkName: s.linkMap[addr.LinkIndex].Name,
		})
	}

	return addrSlice
}

func (s *State) updateLink(update netlink.LinkUpdate) bool {
	attr := update.Link.Attrs()
	link := linkState{Index: attr.Index, Name: attr.Name}

	slog.Debug("link update event", slog.String("link", link.Name))

	oldValue, exists := s.linkMap[link.Index]
	if !exists || oldValue != link {
		s.linkMap[link.Index] = link
		return true
	} else {
		slog.Debug("link did not change", slog.String("link", link.Name))
		return false
	}
}

func (s *State) updateAddr(update netlink.AddrUpdate) bool {
	ipStr := update.LinkAddress.String()
	addr := addrState{
		IPNet:     update.LinkAddress,
		LinkIndex: update.LinkIndex,
	}

	oldValue, exists := s.addrMap[ipStr]

	if update.NewAddr {
		slog.Debug("add address event", slog.String("ip", ipStr))

		if !exists || oldValue.LinkIndex != addr.LinkIndex || !oldValue.IP.Equal(addr.IP) {
			s.addrMap[ipStr] = addr
			return true
		} else {
			slog.Debug("address did not change. skip update", slog.String("ip", ipStr))
		}
	} else {
		slog.Debug("delete address event", slog.String("ip", ipStr))

		if exists {
			delete(s.addrMap, ipStr)
			return true
		} else {
			slog.Debug("address was not in state. skip delete.", slog.String("ip", ipStr))
		}
	}

	return false
}

func (s *State) handleUpdates(lu <-chan netlink.LinkUpdate, au <-chan netlink.AddrUpdate, su chan<- *State) {
	su <- s

	for {
		select {
		case l := <-lu:
			if !s.updateLink(l) {
				continue
			}

		case a := <-au:
			if !s.updateAddr(a) {
				continue
			}
		}

		su <- s
	}
}

func (s *State) Monitor() (<-chan *State, error) {
	var (
		linkUpdates  = make(chan netlink.LinkUpdate, 8)
		addrUpdates  = make(chan netlink.AddrUpdate, 8)
		stateUpdates = make(chan *State)
	)

	go s.handleUpdates(linkUpdates, addrUpdates, stateUpdates)
	return stateUpdates, subscribeNetlink(linkUpdates, addrUpdates)
}
