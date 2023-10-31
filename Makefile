# Makefile
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.PHONY: all
all: ;

# BUILD-CLIENT
.PHONY: build-gclient
build-gclient:
	go build -C ./cmd/gclient -o $(shell pwd)/cmd/gclient/gclient

# BUILD-SERVER
.PHONY: build-gserver
build-gserver:
	go build -C ./cmd/gserver -o $(shell pwd)/cmd/gserver/gserver

# BUILD
.PHONY: build
build:
	make build-gclient
	make build-gserver

# TESTS
.PHONY: tests
tests:
	go vet ./...
	go test ./... --tags=usetempdir -v -race -count=1 -coverpkg=./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o ./coverage.html

# LINT
.PHONY: lint
lint:
	[ -d $(ROOT_DIR)/golangci-lint ] || mkdir -p $(ROOT_DIR)/golangci-lint
	docker run --rm \
    -v $(ROOT_DIR):/app \
    -v $(ROOT_DIR)/golangci-lint/.cache:/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.54 \
        golangci-lint run \
        -c .golangci-lint.yml \
    > ./golangci-lint/report.json

# POSTGRESQL
.PHONY: run-pg
run-pg:
	docker run --rm \
		--name=postgresql \
		-v $(ROOT_DIR)/deployments/db/init/:/docker-entrypoint-initdb.d \
		-v $(ROOT_DIR)/deployments/db/data/:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=gopher \
		-d \
		-p 5432:5432 \
		postgres:15.3

.PHONY: stop-pg
stop-pg:
	docker stop postgresql

.PHONY: clean-data
clean-data:
	rm -rf ./deployments/db/data/	

# MOCKS
.PHONY: mocks
mocks: protoc
	mockgen -source=internal/models/users.go -destination=internal/server/mock_users_service.go -package server
	mockgen -source=internal/models/records.go -destination=internal/server/mock_records_service.go -package server
	mockgen -source=internal/models/users.go -destination=internal/agent/mock_user_storage.go -package agent
	mockgen -source=internal/models/records.go -destination=internal/agent/mock_records_storage.go -package agent
	
# PROTOBUF
.PHONY: protoc
protoc:
	protoc proto/v1/*.proto  --proto_path=proto/v1 \
	--go_out=internal/server --go_opt=module=github.com/ArtemShalinFe/gophkeeper/internal/server \
	--go-grpc_out=internal/server --go-grpc_opt=module=github.com/ArtemShalinFe/gophkeeper/internal/server