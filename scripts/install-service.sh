#!/bin/bash

# GoTunnel Server Service Installation Script
# This script installs tunnel-server as a systemd service

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  GoTunnel Service Installation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root or with sudo${NC}"
    echo "Usage: sudo ./scripts/install-service.sh"
    exit 1
fi

# Check if tunnel-server binary exists
if [ ! -f "bin/server" ]; then
    echo -e "${YELLOW}Binary not found. Building...${NC}"
    go build -o bin/server ./cmd/tunnel-server
fi

# Install binary to /usr/local/bin
echo -e "${GREEN}Installing binary...${NC}"
cp bin/server /usr/local/bin/tunnel-server
chmod +x /usr/local/bin/tunnel-server
echo -e "${GREEN}✓ Binary installed to /usr/local/bin/tunnel-server${NC}"

# Install systemd service file
echo -e "${GREEN}Installing systemd service...${NC}"
cp scripts/tunnel-server.service /etc/systemd/system/
chmod 644 /etc/systemd/system/tunnel-server.service

# Update WorkingDirectory in service file to current directory
CURRENT_DIR=$(pwd)
sed -i "s|WorkingDirectory=/home/ec2-user/goTunnel|WorkingDirectory=$CURRENT_DIR|g" /etc/systemd/system/tunnel-server.service

# Get actual username
ACTUAL_USER=${SUDO_USER:-$USER}
if [ "$ACTUAL_USER" = "root" ]; then
    ACTUAL_USER="ec2-user"
fi

# Update User in service file
sed -i "s|User=ec2-user|User=$ACTUAL_USER|g" /etc/systemd/system/tunnel-server.service
sed -i "s|Group=ec2-user|Group=$ACTUAL_USER|g" /etc/systemd/system/tunnel-server.service

echo -e "${GREEN}✓ Service file installed${NC}"

# Reload systemd
echo -e "${GREEN}Reloading systemd...${NC}"
systemctl daemon-reload
echo -e "${GREEN}✓ Systemd reloaded${NC}"

# Enable service (start on boot)
echo -e "${GREEN}Enabling service...${NC}"
systemctl enable tunnel-server.service
echo -e "${GREEN}✓ Service enabled (will start on boot)${NC}"

# Start service
echo -e "${GREEN}Starting service...${NC}"
systemctl start tunnel-server.service
echo -e "${GREEN}✓ Service started${NC}"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Installation Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Show status
echo -e "${BLUE}Service Status:${NC}"
systemctl status tunnel-server.service --no-pager

echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo -e "  ${YELLOW}sudo systemctl start tunnel-server${NC}    # Start service"
echo -e "  ${YELLOW}sudo systemctl stop tunnel-server${NC}     # Stop service"
echo -e "  ${YELLOW}sudo systemctl restart tunnel-server${NC}  # Restart service"
echo -e "  ${YELLOW}sudo systemctl status tunnel-server${NC}   # Check status"
echo -e "  ${YELLOW}sudo journalctl -u tunnel-server -f${NC}   # View logs (live)"
echo -e "  ${YELLOW}sudo journalctl -u tunnel-server -n 50${NC} # View last 50 log lines"
echo ""
echo -e "${GREEN}Server is now running at: wss://easysource-mortalengine.hirequotient.com/tunnel${NC}"
echo ""
