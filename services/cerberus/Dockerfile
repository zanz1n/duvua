FROM ubuntu

COPY --from=duvua/cargo_build /build/target/release/cerberus /service

CMD [ "/service" ]
