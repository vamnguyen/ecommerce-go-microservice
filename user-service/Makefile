.PHONY: proto build run test clean docker-build deps

deps:
	@echo "Installing dependencies for user-service..."
	@cd $(shell pwd) && go mod download && go mod tidy

proto:
	@echo "Generating proto files..."
	@mkdir -p gen/go
	@PATH="$$(go env GOPATH)/bin:$$PATH" protoc --go_out=gen/go --go_opt=paths=source_relative \
		--go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=gen/go --grpc-gateway_opt=paths=source_relative \
		--proto_path=proto \
		--proto_path=../proto-common \
		proto/user.proto

build:
	@echo "Building user-service..."
	@go build -o bin/user-service ./cmd/server

run:
	@echo "Running user-service..."
	@go run ./cmd/server

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning..."
	@rm -rf bin gen

docker-build:
	@echo "Building Docker image..."
	@docker build -t user-service:latest .

