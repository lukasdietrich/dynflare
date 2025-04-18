package hook

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/lukasdietrich/dynflare/internal/config"
	"github.com/lukasdietrich/dynflare/internal/nameserver"
)

type Hook struct {
	config.Hook
}

func New(cfg config.Hook) *Hook {
	return &Hook{cfg}
}

func (h *Hook) Execute(record nameserver.Record) error {
	if h.Command == "" {
		return nil
	}

	environment := buildEnvironment(record)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		h.Command.StringWith(environment),
		expandArguments(h.Args, environment)...,
	)

	cmd.Env = environment

	slog.Debug("executing hook", slog.Any("args", cmd.Args))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	slog.Debug("hook finished", slog.String("output", string(output)))
	return nil
}

func buildEnvironment(record nameserver.Record) []string {
	var environment []string

	environment = append(environment, os.Environ()...)
	environment = append(environment,
		fmt.Sprintf("ZONE=%s", record.Zone),
		fmt.Sprintf("DOMAIN=%s", record.Domain),
		fmt.Sprintf("KIND=%s", record.Kind),
		fmt.Sprintf("IP=%s", record.IP.String()),
	)

	return environment
}

func expandArguments(args []config.EnvString, environment []string) []string {
	var argsExpanded []string

	for _, arg := range args {
		argsExpanded = append(argsExpanded, arg.StringWith(environment))
	}

	return argsExpanded
}
