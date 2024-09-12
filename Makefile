.PHONY: default

default: build build-dev

run:
	go run cmd/bot/main.go --migrate

run-player:
	go run cmd/player/main.go

build: build-bot build-player

build-bot:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua-bot/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot cmd/bot/main.go

build-player:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua-bot/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player cmd/player/main.go

build-dev: build-dev-bot build-dev-player

build-dev-bot:
	go build \
		-ldflags "-X github.com/zanz1n/duvua-bot/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot-debug cmd/bot/main.go

build-dev-player:
	go build \
		-ldflags "-X github.com/zanz1n/duvua-bot/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player-debug cmd/player/main.go

test:
	go test ./... -v --race
