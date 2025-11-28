#!/bin/bash

# Script to install protoc plugins required for proto generation

set -e

echo "🔧 Installing protoc plugins..."

# Get GOPATH
GOPATH=$(go env GOPATH)
BIN_DIR="$GOPATH/bin"

# Ensure bin directory exists
mkdir -p "$BIN_DIR"

echo "Installing protoc-gen-go..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

echo "Installing protoc-gen-go-grpc..."
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

echo "Installing protoc-gen-grpc-gateway..."
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

echo ""
echo "✅ All protoc plugins installed to: $BIN_DIR"
echo ""
echo "Make sure $BIN_DIR is in your PATH:"
echo "  export PATH=\"\$PATH:$BIN_DIR\""
echo ""
echo "Or add to your ~/.zshrc or ~/.bashrc:"
echo "  export PATH=\"\$(go env GOPATH)/bin:\$PATH\""

