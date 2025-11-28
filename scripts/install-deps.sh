#!/bin/bash

# Script to install dependencies for all services

set -e

echo "🚀 Installing dependencies for all services..."
echo ""

# Get the root directory of the project
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to install deps for a service
install_service_deps() {
    local service_name=$1
    local service_dir="$ROOT_DIR/$service_name"
    
    if [ ! -d "$service_dir" ]; then
        echo "⚠️  Service directory not found: $service_dir"
        return
    fi
    
    if [ ! -f "$service_dir/go.mod" ]; then
        echo "⚠️  go.mod not found in $service_name, skipping..."
        return
    fi
    
    echo -e "${BLUE}📦 Installing dependencies for $service_name...${NC}"
    cd "$service_dir"
    
    if command -v go &> /dev/null; then
        go mod download
        go mod tidy
        echo -e "${GREEN}✅ $service_name dependencies installed${NC}"
    else
        echo "❌ Go is not installed. Please install Go first."
        exit 1
    fi
    
    echo ""
}

# Install dependencies for each service
install_service_deps "auth-service"
install_service_deps "user-service"
install_service_deps "order-service"

echo -e "${GREEN}🎉 All dependencies installed successfully!${NC}"

