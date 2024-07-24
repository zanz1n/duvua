FROM golang:1 AS builder

WORKDIR /build
ENV CGO_ENABLED=0

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

COPY . .

RUN --mount=type=cache,target=/gomod-cache \
    --mount=type=cache,target=/go-cache \
    make build

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/bin/duvua-bot /bin/duvua-bot

CMD [ "/bin/duvua-bot" ]
