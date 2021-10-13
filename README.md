# dynflare

`dynflare` is a tool to automatically update dns records at Cloudflare, when the ip changes.

## How it works

The current ips are determined by asking the os.
Only non-loopback unicast addresses are considered and optionally filtered by a suffix.
The last determined ip is cached at `~/.cache/dynflare/{domain}.{kind}.txt` and compared to the
current ip before doing any api calls to Cloudflare.

## Usage

To test the configuration you can simply run dynflare with an optional flag pointing at the config
file:

```sh
dynflare -config my-config.toml
```

When everything works, you can install and enable a systemd timer to `~/.local/share/systemd/user`
using the following series of commands:

```sh
dynflare -config my-config.toml -install
systemctl --user enable --now dynflare.timer
```

The default interval is one minute.
