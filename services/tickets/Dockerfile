FROM ubuntu

COPY --from=duvua/cargo_build /build/target/release/tickets /service

CMD [ "/service" ]
