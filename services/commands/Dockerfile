FROM ubuntu

COPY --from=duvua/cargo_build /build/target/release/commands /service

CMD [ "/service" ]
