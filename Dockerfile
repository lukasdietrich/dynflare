FROM docker.io/library/golang:alpine as builder

	WORKDIR /github.com/lukasdietrich/dynflare
	COPY . .

	RUN go build ./cmd/dynflare

FROM docker.io/library/alpine

	WORKDIR /app
	COPY --from=builder /github.com/lukasdietrich/dynflare/dynflare .

	VOLUME /data
	USER nobody

	CMD ["/app/dynflare", "-config", "/data/config.toml", "-cache", "/data/cache.toml"]
