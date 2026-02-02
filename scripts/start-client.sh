#!/bin/bash

# GoTunnel Client Start Script
# Usage: ./scripts/start-client.sh <server-url> <local-port>
#
# Examples:
#   ./scripts/start-client.sh ws://localhost:8080/tunnel 3000
#   ./scripts/start-client.sh ws://your-ec2.com:8080/tunnel 5173

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check arguments
if [ $# -lt 2 ]; then
    echo -e "${RED}Error: Missing required arguments${NC}"
    echo ""
    echo "Usage: $0 <server-url> <local-port>"
    echo ""
    echo "Examples:"
    echo "  $0 ws://localhost:8080/tunnel 3000"
    echo "  $0 ws://your-ec2.com:8080/tunnel 5173"
    echo ""
    exit 1
fi

SERVER_URL="$1"
LOCAL_PORT="$2"

# Validate WebSocket URL
if [[ ! "$SERVER_URL" =~ ^wss?:// ]]; then
    echo -e "${RED}Error: Server URL must start with 'ws://' or 'wss://'${NC}"
    echo ""
    echo "You provided: $SERVER_URL"
    echo ""
    echo "Correct examples:"
    echo "  ws://localhost:8080/tunnel"
    echo "  ws://54.123.45.67:8080/tunnel"
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
