FROM golang:1 AS builder

WORKDIR /build
ENV CGO_ENABLED=1

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-bot

FROM gcr.io/distroless/cc-debian12

COPY --from=builder /build/bin/duvua-bot /usr/bin/duvua-bot

ENTRYPOINT [ "/usr/bin/duvua-bot" ]
