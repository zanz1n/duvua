run:
	CGO_ENABLED=1 go run cmd/bot/main.go --migrate

run-player:
	CGO_ENABLED=0 go run cmd/player/main.go

build:
	CGO_ENABLED=1 go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua-bot/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot cmd/bot/main.go

	CGO_ENABLED=0 go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua-bot/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player cmd/player/main.go

build-dev:
	CGO_ENABLED=1 go build \
		-ldflags "-X github.com/zanz1n/duvua-bot/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot-debug cmd/bot/main.go

	CGO_ENABLED=0 go build \
		-ldflags "-X github.com/zanz1n/duvua-bot/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player-debug cmd/player/main.go

test:
	CGO_ENABLED=1 go test ./... -v --race
