#!/bin/bash

# Run all tests with coverage

set -e

echo "Running tests..."

# Run unit tests
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Display coverage summary
go tool cover -func=coverage.out

echo ""
echo "Coverage report generated: coverage.html"
