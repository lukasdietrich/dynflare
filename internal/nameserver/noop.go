package nameserver

import "log/slog"

var (
	_ Nameserver = &noopNameserver{}
)

type noopNameserver struct{}

func (*noopNameserver) UpdateRecord(record Record) error {
	slog.Warn("noop nameserver configured for record", slog.Any("record", record))
	return nil
}
