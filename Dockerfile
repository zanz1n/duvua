FROM golang:1 AS builder

WORKDIR /build
COPY . .

RUN CGO_ENABLED=0 make build

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/bin/duvua-bot /duvua-bot

CMD [ "/duvua-bot" ]
