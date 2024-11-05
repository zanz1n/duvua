FROM golang:1 AS builder

ARG VERSION=""

WORKDIR /build
ENV CGO_ENABLED=1

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
    -o bin/duvua-bot github.com/zanz1n/duvua/cmd/bot

FROM gcr.io/distroless/cc-debian12

COPY --from=builder /build/bin/duvua-bot /usr/bin/duvua-bot

ENTRYPOINT [ "/usr/bin/duvua-bot" ]
