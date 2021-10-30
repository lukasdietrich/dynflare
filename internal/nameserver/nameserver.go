package nameserver

import (
	"errors"
	"fmt"
	"net"

	"github.com/lukasdietrich/dynflare/internal/config"
)

var (
	ErrUnknownZone = errors.New("zone unknown")
)

type RecordKind string

const (
	KindV4 = RecordKind("A")
	KindV6 = RecordKind("AAAA")
)

type Record struct {
	Zone   string
	Domain string
	Kind   RecordKind
	IP     net.IP
}

type Nameserver interface {
	UpdateRecord(record Record) error
}

func New(cfg config.Nameserver) (Nameserver, error) {
	switch cfg.Provider {
	case "cloudflare":
		return newCloudflare(cfg.Credentials)
	default:
		return nil, fmt.Errorf("unknown provider %q", cfg.Provider)
	}
}
