#!/bin/bash

# GoTunnel Server Start Script
# Usage: ./scripts/start-server.sh [port] [domain]
#
# Examples:
#   ./scripts/start-server.sh                    # Start on default port 8080
#   ./scripts/start-server.sh 9090               # Start on port 9090
#   ./scripts/start-server.sh 8080 tunnel.com    # With custom domain

set -e

# Default values
PORT="${1:-8080}"
DOMAIN="${2:-}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}    GoTunnel Server Start Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if binary exists
if [ ! -f "bin/server" ]; then
    echo -e "${YELLOW}Server binary not found. Building...${NC}"
    mkdir -p bin
    go build -o bin/server ./cmd/server
    echo -e "${GREEN}✓ Build complete${NC}"
    echo ""
fi

# Build the command
CMD="./bin/server --port $PORT"
if [ -n "$DOMAIN" ]; then
    CMD="$CMD --domain $DOMAIN"
fi

# Display configuration
echo -e "${GREEN}Configuration:${NC}"
echo -e "  Port: ${YELLOW}$PORT${NC}"
if [ -n "$DOMAIN" ]; then
    echo -e "  Domain: ${YELLOW}$DOMAIN${NC}"
fi
echo ""

# Start the server
echo -e "${GREEN}Starting GoTunnel server...${NC}"
echo -e "${BLUE}Press Ctrl+C to stop${NC}"
echo ""

exec $CMD
