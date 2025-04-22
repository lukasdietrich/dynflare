package nameserver

import (
	"errors"
	"fmt"
	"net"

	"github.com/lukasdietrich/dynflare/internal/config"
)

type RecordKind string

const (
	KindV4 = RecordKind("A")
	KindV6 = RecordKind("AAAA")
)

type Record struct {
	Zone    string
	Domain  string
	Kind    RecordKind
	IP      net.IP
	Comment string
}

type Nameserver interface {
	UpdateRecord(record Record) error
}

func New(cfg config.Nameserver) (Nameserver, error) {
	switch cfg.Provider {
	case "cloudflare":
		return newCloudflare(cfg.Credentials.String())
	case "noop":
		return &noopNameserver{}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q", cfg.Provider)
	}
}

type permanentClientError struct {
	cause error
}

func (e *permanentClientError) Error() string {
	return e.cause.Error()
}

func (e *permanentClientError) Unwrap() error {
	return e.cause
}

func wrapPermanentClientError(err error) error {
	return &permanentClientError{cause: err}
}

func IsPermanentClientError(err error) bool {
	var pcErr *permanentClientError
	return errors.As(err, &pcErr)

}
