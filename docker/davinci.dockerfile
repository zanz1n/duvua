FROM golang:1 AS builder

ARG VERSION
ARG DEBUG=0

WORKDIR /build
ENV CGO_ENABLED=1
ENV OUTPUT=/build/bin/duvua-davinci
ENV GOFLAGS=-buildvcs=false

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build-davinci

FROM gcr.io/distroless/cc-debian12

COPY --from=builder /build/bin/duvua-davinci /usr/bin/duvua-davinci

ENTRYPOINT [ "/usr/bin/duvua-davinci" ]

EXPOSE 8080
