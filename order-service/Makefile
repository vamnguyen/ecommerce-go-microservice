.PHONY: proto build run test clean docker-build deps dev

deps:
	@echo "Installing dependencies for order-service..."
	@cd $(shell pwd) && go mod download && go mod tidy

proto:
	@echo "Generating proto files..."
	@mkdir -p gen/go
	@PATH="$$(go env GOPATH)/bin:$$PATH" protoc --go_out=gen/go --go_opt=paths=source_relative \
		--go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=gen/go --grpc-gateway_opt=paths=source_relative \
		--proto_path=proto \
		--proto_path=../proto-common \
		proto/order.proto

build:
	@echo "Building order-service..."
	@go build -o bin/order-service ./cmd/server

run:
	@echo "Running order-service..."
	@go run ./cmd/server

dev:
	$(shell go env GOPATH)/bin/air

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin gen

docker-build:
	@echo "Building Docker image..."
	@docker build -t order-service:latest .

