FROM rust:1.72

WORKDIR /root

RUN cargo install sqlx-cli

CMD [ "bash" ]
