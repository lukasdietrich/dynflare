package dyndns

import (
	"fmt"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/rs/zerolog/log"

	"github.com/lukasdietrich/dynflare/internal/config"
)

type notifier struct {
	router *router.ServiceRouter
}

func newNotifier(cfg config.Config) (*notifier, error) {
	if len(cfg.Notifications) == 0 {
		return nil, nil
	}

	var urls []string

	for _, notification := range cfg.Notifications {
		urls = append(urls, notification.URL)
	}

	router, err := shoutrrr.CreateSender(urls...)
	if err != nil {
		return nil, fmt.Errorf("could not create shoutrrr sender: %w", err)
	}

	return &notifier{router}, nil
}

func (n *notifier) notify(format string, v ...interface{}) {
	if n == nil {
		log.Debug().Msg("no notification urls configured")
		return
	}

	message := fmt.Sprintf(format, v...)

	log.Debug().
		Str("content", message).
		Msg("sending notification")

	if errs := filterEmptyErrors(n.router.Send(message, nil)); len(errs) > 0 {
		log.Warn().
			Errs("errors", errs).
			Msg("could not send notification")
	}
}

func filterEmptyErrors(errs []error) []error {
	var nonEmptyErrors []error

	for _, err := range errs {
		if err != nil {
			nonEmptyErrors = append(nonEmptyErrors, err)
		}
	}

	return nonEmptyErrors
}
