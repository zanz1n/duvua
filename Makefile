run:
	go run cmd/bot/main.go --migrate

build:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua-bot/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot cmd/bot/main.go

build-dev:
	go build \
		-ldflags "-X github.com/zanz1n/duvua-bot/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot-debug cmd/bot/main.go

test:
	go test ./... -v --race
