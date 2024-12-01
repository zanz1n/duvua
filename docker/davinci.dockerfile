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
    -buildvcs=false -o bin/duvua-davinci github.com/zanz1n/duvua/cmd/davinci

FROM gcr.io/distroless/cc-debian12

COPY --from=builder /build/bin/duvua-davinci /usr/bin/duvua-davinci

ENTRYPOINT [ "/usr/bin/duvua-davinci" ]

EXPOSE 8080
