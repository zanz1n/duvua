FROM golang:1 AS builder

WORKDIR /build
ENV CGO_ENABLED=0

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-player

FROM alpine

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

COPY --from=builder /build/bin/duvua-player /bin/duvua-player

CMD [ "/bin/duvua-player" ]
