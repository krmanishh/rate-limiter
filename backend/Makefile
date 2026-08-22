BINARY_NAME := rate-limiter-server
BUILD_DIR   := bin
GOLANGCI_LINT := $(shell go env GOPATH)/bin/golangci-lint

.PHONY: build run test race vet fmt lint docker-up docker-down clean

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	$(GOLANGCI_LINT) run ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

clean:
	rm -rf $(BUILD_DIR)
	go clean
