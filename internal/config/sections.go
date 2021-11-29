package config

type Log struct {
	// Level is the log level (eg. "debug", "info", "warn" or "error").
	Level string `toml:"level"`
	// Format is the log output format.
	// "json" emits machine readable json objects, "text" emits human readable text.
	Format string `toml:"format"`
	// Caller is a flag to enable source filenames and line numbers in the log output.
	Caller bool `toml:"caller"`
	// Timestamp is a flag to enable timestamps in log the output.
	Timestamp bool `toml:"timestamp"`
}

type Notification struct {
	// Shoutrrr URL (see <https://github.com/containrrr/shoutrrr>)
	URL string `toml:"url"`
}

type Nameserver struct {
	// Provider is the name of the api adapter used to change dns records.
	Provider string `toml:"provider"`
	// Credentials is a provider dependant secret string (eg. a Cloudflare API Token).
	Credentials string `toml:"credentials"`
	// Zones is a slice of zones this nameserver is responsible for (eg. "example.com").
	Zones []string `toml:"zones"`
}

type Domain struct {
	// Name is the domain name including its zone (eg. "raspberry-pi.example.com").
	Name string `toml:"name"`
	// Zone is the zone name matching a configured nameserver.
	Zone string `toml:"zone"`
	// Kind is the dns record type (eg. "AAAA" for IPv6 or "A" for IPv4).
	Kind string `toml:"kind"`
	// Interface is an optional filter for the network interface (eg. wlan0 or eth0).
	Interface string `toml:"interface"`
	// Suffix is an optional filter for the part after the network mask
	// (eg. the remaining 64 bits of a ::/64 address).
	Suffix string `toml:"suffix"`
}
