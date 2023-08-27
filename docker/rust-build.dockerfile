FROM rust:1.72

WORKDIR /build

COPY . .

RUN cargo build --release
