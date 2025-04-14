from docker.io/library/golang:alpine as builder

	workdir /github.com/lukasdietrich/dynflare
	copy internal ./internal
	copy cmd ./cmd
	copy go* .

	run ls -l
	run go build -v ./cmd/dynflare

from docker.io/library/alpine

	workdir /app
	copy --from=builder /github.com/lukasdietrich/dynflare/dynflare .
	copy LICENSE .

	label org.opencontainers.image.authors="Lukas Dietrich <lukas@lukasdietrich.com>"
	label org.opencontainers.image.source="https://github.com/lukasdietrich/dynflare"

	volume /data
	user nobody

	cmd ["/app/dynflare", "-config", "/data/config.toml", "-cache", "/data/cache.toml"]
