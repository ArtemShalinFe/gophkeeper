# Makefile
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: all
all: ;

# BUILD-CLIENT
.PHONY: build-client
buildclient:
	go build -C ./cmd/client -o $(shell pwd)/cmd/client/gophermart

# BUILD-AGENT
.PHONY: build-agent
build-agent:
	go build -C ./cmd/agent -o $(shell pwd)/cmd/agent/gophermart

# BUILD-SERVER
.PHONY: build-server
build-server:
	go build -C ./cmd/server -o $(shell pwd)/cmd/server/gophermart

# BUILD
.PHONY: build
build:
	build-agent
	build-client
	build-server

# TESTS
.PHONY: tests
tests:
	go vet ./...
	go test ./... --tags=usetempdir -v -race -count=1 -coverpkg=./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o ./coverage.html

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

.PHONY: mocks
mocks: protoc
	mockgen -source=internal/metrics/metrics.go -destination=internal/metrics/mock_metrics.go -package metrics
	mockgen -source=internal/metcoll/handlers.go -destination=internal/metcoll/mock_handlers.go -package metcoll
	mockgen -source=internal/metcoll/metcoll_grpc.pb.go -destination=internal/metcoll/mock_grpc_pb.go -package metcoll

.PHONY: protoc
protoc:
	protoc proto/v1/*.proto  --proto_path=proto/v1 \
	--go_out=internal/server --go_opt=module=github.com/ArtemShalinFe/gophkeeper/internal/server \
	--go-grpc_out=internal/server --go-grpc_opt=module=github.com/ArtemShalinFe/gophkeeper/internal/server