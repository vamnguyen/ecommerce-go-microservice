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

.DEFAULT_GOAL := help
