FROM golang:1 AS builder

ARG VERSION
ARG DEBUG=0

WORKDIR /build
ENV CGO_ENABLED=0
ENV OUTPUT=/build/bin/duvua-bot
ENV GOFLAGS=-buildvcs=false

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-bot

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/bin/duvua-bot /usr/bin/duvua-bot

ENTRYPOINT [ "/usr/bin/duvua-bot" ]
