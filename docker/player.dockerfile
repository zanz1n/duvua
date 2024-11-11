FROM golang:1 AS builder

ARG VERSION=""

WORKDIR /build
ENV CGO_ENABLED=0

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    if [ -z ${VERSION} ]; then \
    VERSION_TAG=release-`git rev-parse --short HEAD`; \
    else \
    VERSION_TAG=${VERSION}; \
    fi; \
    go build \
    -ldflags "-s -w -X github.com/zanz1n/duvua/config.Version=${VERSION_TAG}" \
    -o bin/duvua-player ./cmd/player

FROM alpine:3

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

COPY --from=builder /build/bin/duvua-player /usr/bin/duvua-player

ENTRYPOINT [ "/usr/bin/duvua-player" ]

EXPOSE 8080
