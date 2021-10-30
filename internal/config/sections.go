package config

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
