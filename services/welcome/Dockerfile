FROM ubuntu

COPY --from=duvua/cargo_build /build/target/release/welcome /service

CMD [ "/service" ]
