FROM ubuntu

COPY --from=duvua/cargo_build /build/target/release/chimera /service

CMD [ "/service" ]
