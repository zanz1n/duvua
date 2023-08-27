#!/bin/sh

docker build --tag duvua/cargo_build --file ./docker/rust-build.dockerfile .
docker-compose pull
docker-compose build --parallel
