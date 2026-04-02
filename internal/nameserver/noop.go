package nameserver

import (
	"context"
	"log/slog"
)

var (
	_ Nameserver = &noopNameserver{}
)

type noopNameserver struct{}

func (*noopNameserver) UpdateRecord(_ context.Context, record Record) (bool, error) {
	slog.Warn("noop nameserver configured for record", slog.Any("record", record))
	return false, nil
}
