ARG CGO_ENABLED=0
ARG VERSION
ARG DEBUG=0
ARG SERVICE
ARG IMAGE=raw-cc${CGO_ENABLED}

FROM alpine:3 AS img-ffmpeg

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

FROM gcr.io/distroless/cc-debian12 AS img-raw-cc1
FROM gcr.io/distroless/static-debian12 AS img-raw-cc0

FROM golang:bookworm AS img-golang

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

RUN apt-get update -y
RUN apt-get install -y unzip

FROM img-golang AS builder

ARG CGO_ENABLED
ARG VERSION
ARG DEBUG
ARG SERVICE

WORKDIR /build
ENV OUTPUT=/build/bin/duvua-${SERVICE}
ENV GOFLAGS=-buildvcs=false

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-${SERVICE}

FROM img-${IMAGE}

ARG SERVICE

COPY --from=builder /build/bin/duvua-${SERVICE} /usr/bin/service

ENTRYPOINT [ "/usr/bin/service" ]

EXPOSE 8080
