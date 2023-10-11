# Makefile
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: all
all: ;

# BUILD-CLIENT
.PHONY: build-client
build:
	go build -C ./cmd/client -o $(shell pwd)/cmd/client/gophermart

# BUILD-SERVER
.PHONY: build-server
build:
	go build -C ./cmd/server -o $(shell pwd)/cmd/client/gophermart

# BUILD
.PHONY: build
build: 
	build-client
	build-server

# TESTS
.PHONY: tests
tests:
	go test ./... -v -race

.PHONY: lint
lint:
	[ -d $(ROOT_DIR)/golangci-lint ] || mkdir -p $(ROOT_DIR)/golangci-lint
	docker run --rm \
    -v $(ROOT_DIR):/app \
    -v $(ROOT_DIR)/golangci-lint/.cache:/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.53.3 \
        golangci-lint run \
        -c .golangci-lint.yml \
    > ./golangci-lint/report.json
