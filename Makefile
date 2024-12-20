.PHONY: default

default: check build build-dev

run: build-dev-bot
	./bin/duvua-bot-debug --migrate

run-player: build-dev-player
	./bin/duvua-player-debug

run-davinci: build-dev-davinci
	./bin/duvua-davinci-debug

build: build-bot build-player build-davinci

build-bot:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot github.com/zanz1n/duvua/cmd/bot

build-player:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player github.com/zanz1n/duvua/cmd/player

build-davinci:
	go build \
		-ldflags "-s -w -X github.com/zanz1n/duvua/config.Version=release-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-davinci github.com/zanz1n/duvua/cmd/davinci

build-dev: build-dev-bot build-dev-player build-dev-davinci

build-dev-bot:
	go build \
		-ldflags "-X github.com/zanz1n/duvua/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-bot-debug github.com/zanz1n/duvua/cmd/bot

build-dev-player:
	go build \
		-ldflags "-X github.com/zanz1n/duvua/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-player-debug github.com/zanz1n/duvua/cmd/player

build-dev-davinci:
	go build \
		-ldflags "-X github.com/zanz1n/duvua/config.Version=debug-$(shell git rev-parse --short HEAD)" \
		-o bin/duvua-davinci-debug github.com/zanz1n/duvua/cmd/davinci

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

compose-up:
	docker compose pull
	docker compose build --parallel
	docker compose up -d
	docker compose logs -f

compose-down:
	docker compose down

compose-clean:
	docker compose down
	sudo rm -rf .docker-volumes	
