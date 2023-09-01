#!/bin/sh

docker build --tag duvua/cargo_build --file ./docker/rust-build.dockerfile .
docker-compose pull
docker-compose build duvua_migrator
docker-compose up -d duvua_postgresql
sleep 3
docker-compose run --rm -it duvua_migrator bash -c "cargo sqlx database create && cargo sqlx migrate run"
docker-compose down

docker-compose build --parallel
