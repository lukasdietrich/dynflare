package monitor

import (
	"log"
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

func (s *State) updateLink(update netlink.LinkUpdate) {
	attr := update.Link.Attrs()
	link := linkState{Index: attr.Index, Name: attr.Name}

	log.Printf("link update: %q", link.Name)
	s.linkMap[link.Index] = link
}

func (s *State) updateAddr(update netlink.AddrUpdate) {
	ipStr := update.LinkAddress.String()
	addr := addrState{
		IPNet:     update.LinkAddress,
		LinkIndex: update.LinkIndex,
	}

	if update.NewAddr {
		log.Printf("add addr: %q", ipStr)
		s.addrMap[ipStr] = addr
	} else {
		log.Printf("delete addr: %q", ipStr)
		delete(s.addrMap, ipStr)
	}
}

func (s *State) handleUpdates(lu <-chan netlink.LinkUpdate, au <-chan netlink.AddrUpdate, su chan<- *State) {
	for {
		su <- s

		select {
		case l := <-lu:
			s.updateLink(l)

		case a := <-au:
			s.updateAddr(a)
		}
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
