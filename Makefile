ifneq ($(wildcard .env),)
include .env
endif

SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset

DEBUG ?= 0

PREFIX ?= duvua-
SUFIX ?=

BINS = bot davinci player
DIR ?= bin

GO ?= go

VERSION ?= release-$(shell git rev-parse HEAD | head -c8)

GOMODPATH := github.com/zanz1n/duvua
LDFLAGS := -X $(GOMODPATH)/config.Version=$(VERSION)

ifeq ($(DEBUG), 1)
SUFIX += -debug
else
LDFLAGS += -s -w
endif

OS := $(if $(GOOS),$(GOOS),$(shell GOTOOLCHAIN=local $(GO) env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell GOTOOLCHAIN=local $(GO) env GOARCH))

ifeq ($(OS), windows)
SUFIX += .exe
endif

default: check all

all: $(addprefix build-, $(BINS))

run-%: build-%
ifneq ($(OS), $(shell GOTOOLCHAIN=local $(GO) env GOOS))
	$(error when running GOOS must be equal to the current os)
else ifneq ($(ARCH), $(shell GOTOOLCHAIN=local $(GO) env GOARCH))
	$(error when running GOARCH must be equal to the current cpu arch)
else ifneq ($(OUTPUT),)
	$(OUTPUT)
else
	$(DIR)/$(PREFIX)$*-$(OS)-$(ARCH)$(SUFIX)
endif

build-%: $(DIR)
ifneq ($(OUTPUT),)
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags "$(LDFLAGS)" \
	-o $(OUTPUT) $(GOMODPATH)/cmd/$*
else
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags "$(LDFLAGS)" \
	-o $(DIR)/$(PREFIX)$*-$(OS)-$(ARCH)$(SUFIX) $(GOMODPATH)/cmd/$*
endif
ifneq ($(POST_BUILD_CHMOD),)
	chmod $(POST_BUILD_CHMOD) $(DIR)/$(PREFIX)$*-$(OS)-$(ARCH)$(SUFIX)
endif

$(DIR):
	mkdir $(DIR)

TESTFLAGS = -v -race

ifeq ($(SHORTTESTS), 1)
TESTFLAGS += -short
endif

ifeq ($(NOTESTCACHE), 1)
TESTFLAGS += -count=1
endif

test:
ifneq ($(SKIPTESTS), 1)
	$(GO) test ./... $(TESTFLAGS)
else
    $(warning skipped tests)
endif

.SILENT: gen-checksums
gen-checksums: $(DIR)
	checksum=""; \
	for filename in $(DIR)/*; do \
		checksum+=$$(cd $(DIR) && sha256sum $${filename#"$(DIR)/"}); \
		checksum+="\n"; \
	done; \
	echo -e "\n#### SHA256 Checksum\n\`\`\`\n$$checksum\`\`\`" >> ./RELEASE_CHANGELOG; \
    echo -e "$$checksum" > checksums.txt;

check: generate test

update:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install github.com/srikrsna/protoc-gen-gotag@latest
	$(GO) mod tidy
	$(GO) get -u ./...
	$(GO) mod tidy

generate:
	protoc -I $(shell $(GO) env GOMODCACHE)/github.com/srikrsna/protoc-gen-gotag@* \
		-I . --go_out=./pkg/pb --go-grpc_out=./pkg/pb ./api/proto/*/*.proto

	protoc -I $(shell $(GO) env GOMODCACHE)/github.com/srikrsna/protoc-gen-gotag@* \
		-I . --gotag_out=outdir="./pkg/pb":. ./api/proto/*/*.proto

fmt:
	go fmt ./...

docker-%:
	docker buildx build -f docker/$*.dockerfile -t duvua-$* \
	--build-arg VERSION=$(VERSION) --build-arg DEBUG=$(DEBUG) .

ifneq ($(shell which docker-compose),)
DOCKER_COMPOSE := docker-compose
else
DOCKER_COMPOSE := docker compose
endif

compose-up:
	$(DOCKER_COMPOSE) pull
	$(DOCKER_COMPOSE) build --parallel
	$(DOCKER_COMPOSE) up -d
	$(DOCKER_COMPOSE) logs -f

compose-down:
	$(DOCKER_COMPOSE) down

compose-clean:
	$(DOCKER_COMPOSE) down
	sudo rm -rf .docker-volumes

debug:
	@echo DEBUG = $(DEBUG)
	@echo DIR = $(DIR)
	@echo BINNAME = $(PREFIX)%-$(OS)-$(ARCH)$(SUFIX)
	@echo GOMODPATH = $(GOMODPATH)
	@echo VERSION = $(VERSION)
	@echo BINS = $(BINS)
	@echo LDFLAGS = $(LDFLAGS)
