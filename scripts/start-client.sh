#!/bin/bash

# GoTunnel Client Start Script
# Usage: ./scripts/start-client.sh [local-port] [server-url]
#
# Examples:
#   ./scripts/start-client.sh                    # Uses defaults (port 3000, production server)
#   ./scripts/start-client.sh 8000               # Custom port, production server
#   ./scripts/start-client.sh 3000 ws://localhost:8080/tunnel  # Local testing

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default values
DEFAULT_SERVER="wss://easysource-mortalengine.hirequotient.com/tunnel"
DEFAULT_PORT="3000"

# Parse arguments
LOCAL_PORT="${1:-$DEFAULT_PORT}"
SERVER_URL="${2:-$DEFAULT_SERVER}"

# Validate WebSocket URL
if [[ ! "$SERVER_URL" =~ ^wss?:// ]]; then
    echo -e "${RED}Error: Server URL must start with 'ws://' or 'wss://'${NC}"
    echo ""
    echo "You provided: $SERVER_URL"
    echo ""
    echo "Correct examples:"
    echo "  ws://localhost:8080/tunnel"
    echo "  wss://tunnel.example.com/tunnel"
    echo ""
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}    GoTunnel Client Start Script${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if binary exists
if [ ! -f "bin/client" ]; then
    echo -e "${YELLOW}Client binary not found. Building...${NC}"
    mkdir -p bin
    go build -o bin/client ./cmd/client
    echo -e "${GREEN}✓ Build complete${NC}"
    echo ""
fi

# Check if local port is accessible (optional warning)
if ! nc -z localhost "$LOCAL_PORT" 2>/dev/null; then
    echo -e "${YELLOW}⚠ Warning: No service detected on localhost:$LOCAL_PORT${NC}"
    echo -e "${YELLOW}  Make sure your local app is running on port $LOCAL_PORT${NC}"
    echo ""
fi

# Display configuration
echo -e "${GREEN}Configuration:${NC}"
echo -e "  Server URL: ${YELLOW}$SERVER_URL${NC}"
echo -e "  Local Port: ${YELLOW}$LOCAL_PORT${NC}"
echo ""

# Start the client
echo -e "${GREEN}Starting GoTunnel client...${NC}"
echo -e "${BLUE}Press Ctrl+C to stop${NC}"
echo ""

exec ./bin/client --server "$SERVER_URL" --local-port "$LOCAL_PORT"
