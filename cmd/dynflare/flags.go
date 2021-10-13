package main

import "flag"

type Flags struct {
	ConfigFilename    string
	InstallSystemUnit bool
}

func parseFlags() (f Flags) {
	flag.StringVar(&f.ConfigFilename,
		"config",
		"config.toml",
		"Path to config.toml")

	flag.BoolVar(&f.InstallSystemUnit,
		"install",
		false,
		"Install systemd unit-files to ~/.local/share/systemd/user")

	flag.Parse()
	return
}
