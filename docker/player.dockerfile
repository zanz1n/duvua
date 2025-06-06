FROM golang:1 AS builder

ARG VERSION
ARG DEBUG=0

WORKDIR /build
ENV CGO_ENABLED=0
ENV OUTPUT=/build/bin/duvua-player
ENV GOFLAGS=-buildvcs=false

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-player

FROM alpine:3

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

COPY --from=builder /build/bin/duvua-player /usr/bin/duvua-player

ENTRYPOINT [ "/usr/bin/duvua-player" ]

EXPOSE 8080
