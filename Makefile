.PHONY: help deps deps-all proto proto-all build build-all clean clean-all

help:
	@echo "E-commerce Go Microservice - Makefile"
	@echo ""
	@echo "Available commands:"
	@echo "  make deps              - Install dependencies for all services"
	@echo "  make proto             - Generate proto files for all services"
	@echo "  make build             - Build all services"
	@echo "  make clean             - Clean build artifacts for all services"
	@echo "  make install-protoc    - Install protoc plugins (protoc-gen-go, etc.)"
	@echo ""
	@echo "Service-specific commands:"
	@echo "  make deps-auth         - Install dependencies for auth-service"
	@echo "  make deps-user         - Install dependencies for user-service"
	@echo "  make deps-order        - Install dependencies for order-service"
	@echo ""
	@echo "  make proto-auth        - Generate proto for auth-service"
	@echo "  make proto-user        - Generate proto for user-service"
	@echo "  make proto-order       - Generate proto for order-service"

# Install dependencies
deps: deps-all

deps-all:
	@echo "Installing dependencies for all services..."
	@./scripts/install-deps.sh

deps-auth:
	@cd auth-service && make deps

deps-user:
	@cd user-service && make deps

deps-order:
	@cd order-service && make deps

# Generate proto files
proto: proto-all

proto-all: proto-auth proto-user proto-order

proto-auth:
	@cd auth-service && make proto

proto-user:
	@cd user-service && make proto

proto-order:
	@cd order-service && make proto

# Build all services
build: build-all

build-all: build-auth build-user build-order

build-auth:
	@cd auth-service && make build

build-user:
	@cd user-service && make build

build-order:
	@cd order-service && make build

# Clean all services
clean: clean-all

clean-all: clean-auth clean-user clean-order

clean-auth:
	@cd auth-service && make clean

clean-user:
	@cd user-service && make clean

clean-order:
	@cd order-service && make clean

# Install protoc plugins
install-protoc:
	@./scripts/install-protoc-plugins.sh

.DEFAULT_GOAL := help

