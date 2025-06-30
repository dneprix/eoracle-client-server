# Makefile for Eoracle Client-Server Application

.DEFAULT_GOAL := help

## build: Build binaries
.PHONY: build
build:
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

## build-amd64: Build binaries for macOS AMD64
.PHONY: build-amd64
build-amd64: 
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -o bin/server ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -o bin/client ./cmd/client

## build-arm64: Build binaries for macOS ARM64
.PHONY: build-arm64
build-arm64: 
	@mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -o bin/server ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o bin/client ./cmd/client

## rabbitmq-start: Start RabbitMQ container
.PHONY: rabbitmq-start
rabbitmq-start:
	docker run -d --name rabbitmq-dev \
		-p 5672:5672 \
		-p 15672:15672 \
		-e RABBITMQ_DEFAULT_USER=guest \
		-e RABBITMQ_DEFAULT_PASS=guest \
		rabbitmq:3-management

## rabbitmq-stop: Stop RabbitMQ container
.PHONY: rabbitmq-stop
rabbitmq-stop:
	docker stop rabbitmq-dev || true
	docker rm rabbitmq-dev || true

## test: Run tests
.PHONY: test
test:
	go test ./...

## test-race: Run tests with race detector
.PHONY: test-race
test-race:
	go test -race -v ./...

## test: Run tests with coverage
.PHONY: test-coverage
test-coverage:
	go test -cover ./...
	
## benchmark: Run benchmarks
.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem ./...

## help: Show this help message
.PHONY: help
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

