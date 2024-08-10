run:
	go run cmd/duvua-bot/main.go --migrate

build:
	go build -ldflags "-s -w" -o bin/duvua-bot cmd/duvua-bot/main.go

build-dev:
	go build -o bin/duvua-bot-debug cmd/duvua-bot/main.go

test:
	go test ./... -v --race
