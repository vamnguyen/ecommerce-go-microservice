.PHONY: help run dev build test clean docker-up docker-down migrate

help:
	@echo "Available commands:"
	@echo "  make run          - Run the application"
	@echo "  make dev          - Run with Air (live-reload)"
	@echo "  make build        - Build the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo "  make deps         - Download dependencies"

run:
	go run cmd/server/main.go

dev:
	$(shell go env GOPATH)/bin/air

build:
	go build -o bin/auth-service cmd/server/main.go

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/ tmp/
	rm -f coverage.out coverage.html

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-build:
	docker-compose build

docker-logs:
	docker-compose logs -f auth-service

deps:
	go mod download
	go mod tidy

gen-proto:
	export PATH="$PATH:$(go env GOPATH)/bin"
	source ~/.zshrc
	protoc \
       --go_out=gen/go \
       --go_opt=paths=source_relative \
       --go-grpc_out=gen/go \
       --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=gen/go \
       --grpc-gateway_opt=paths=source_relative \
       --grpc-gateway_opt=logtostderr=true \
       --proto_path=proto \
       proto/auth.proto

.DEFAULT_GOAL := help
