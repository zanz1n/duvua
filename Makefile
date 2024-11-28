.PHONY: default

default: check build build-dev

run: build-dev-bot
	./bin/duvua-bot-debug --migrate

run-player: build-dev-player
	./bin/duvua-player-debug

build: build-bot build-player

build-bot:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot github.com/zanz1n/duvua/cmd/bot

build-player:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player github.com/zanz1n/duvua/cmd/player

build-dev: build-dev-bot build-dev-player

build-dev-bot:
	go build \
		-ldflags "-X github.com/zanz1n/duvua/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot-debug github.com/zanz1n/duvua/cmd/bot

build-dev-player:
	go build \
		-ldflags "-X github.com/zanz1n/duvua/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player-debug github.com/zanz1n/duvua/cmd/player

test:
	go test ./... -v --race

check: generate test

update:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/srikrsna/protoc-gen-gotag@latest
	go get -u ./...
	go mod tidy

generate:
	protoc -I $(shell go env GOMODCACHE)/github.com/srikrsna/protoc-gen-gotag@* \
		-I . --go_out=./pkg/pb --go-grpc_out=./pkg/pb ./api/proto/*/*.proto
	
	protoc -I $(shell go env GOMODCACHE)/github.com/srikrsna/protoc-gen-gotag@* \
		-I . --gotag_out=outdir="./pkg/pb":. ./api/proto/*/*.proto
