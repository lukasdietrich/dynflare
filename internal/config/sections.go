package config

import (
	"fmt"
	"os"
	"strings"
)

var (
	_ fmt.Stringer = EnvString("")
)

// EnvString is a string that may contain environmental variables `${env}` or `$env`.
type EnvString string

// StringWith expands the environmental variables using the current process variables and
// all additionally supplied variables.
func (e EnvString) StringWith(environment []string) string {
	return os.Expand(string(e), func(key string) string {
		for i := len(environment) - 1; i >= 0; i-- {
			if variable := environment[i]; strings.HasPrefix(variable, key+"=") {
				return variable[len(key)+1:]
			}
		}

		return os.Getenv(key)
	})
}

// String expands the environmental variables using the current process variables.
func (e EnvString) String() string {
	return e.StringWith(nil)
}

type Log struct {
	// Level is the log level (eg. "debug", "info", "warn" or "error").
	Level EnvString `toml:"level"`
	// Format is the log output format.
	// "json" emits machine readable json objects, "text" emits human readable text.
	Format EnvString `toml:"format"`
	// Caller is a flag to enable source filenames and line numbers in the log output.
	Caller bool `toml:"caller"`
}

type Notification struct {
	// Shoutrrr URL (see <https://github.com/containrrr/shoutrrr>)
	URL EnvString `toml:"url"`
}

type Nameserver struct {
	// Provider is the name of the api adapter used to change dns records.
	Provider EnvString `toml:"provider"`
	// Credentials is a provider dependant secret string (eg. a Cloudflare API Token).
	Credentials EnvString `toml:"credentials"`
	// Zones is a slice of zones this nameserver is responsible for (eg. "example.com").
	Zones []EnvString `toml:"zones"`
}

type Domain struct {
	// Name is the domain name including its zone (eg. "raspberry-pi.example.com").
	Name EnvString `toml:"name" expr:"name"`
	// Zone is the zone name matching a configured nameserver.
	Zone EnvString `toml:"zone" expr:"zone"`
	// Filter is an expression (https://expr-lang.org/docs/language-definition) to
	// select potential ip candidates.
	Filter EnvString `toml:"filter" expr:"-"`
	// Comment is optionally used to identify the dns record. This way multiple instances can update
	// records for the same domain name.
	Comment EnvString `toml:"comment"`
	// PostHook is a program to execute after an update.
	PostUp Hook `toml:"post-up"`
}

type Hook struct {
	Command EnvString   `toml:"cmd"`
	Args    []EnvString `toml:"args"`
}
