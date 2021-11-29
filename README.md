# dynflare

`dynflare` is a tool to automatically update dns records at Cloudflare, when the ip changes.
Instead of polling external services, `dynflare` relies on kernel events to get the current addresses
(see <https://en.wikipedia.org/wiki/Netlink>).

## TODO Usage
