#!/bin/bash

# GoTunnel - Install from GitHub Release
# This script installs the downloaded release binaries as a systemd service

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  GoTunnel Release Installation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root or with sudo${NC}"
    echo "Usage: sudo ./install-from-release.sh"
    exit 1
fi

# Check if tunnel-server binary exists in current directory
if [ ! -f "tunnel-server" ]; then
    echo -e "${RED}Error: tunnel-server binary not found in current directory${NC}"
    echo ""
    echo "Please extract the release archive first:"
    echo "  tar -xzf goTunnel_*.tar.gz"
    echo "  cd goTunnel_*"
    echo "  sudo ./scripts/install-from-release.sh"
    exit 1
fi

# Get actual username
ACTUAL_USER=${SUDO_USER:-$USER}
if [ "$ACTUAL_USER" = "root" ]; then
    ACTUAL_USER="ec2-user"
fi

echo -e "${GREEN}Installing GoTunnel Server...${NC}"
echo ""

# 1. Install binary
echo -e "${BLUE}[1/4]${NC} Installing binary to /usr/local/bin..."
cp tunnel-server /usr/local/bin/
chmod +x /usr/local/bin/tunnel-server
echo -e "${GREEN}✓ Binary installed${NC}"

# 2. Create systemd service file
echo -e "${BLUE}[2/4]${NC} Creating systemd service..."

if [ -f "scripts/tunnel-server.service" ]; then
    # Use included service file
    cp scripts/tunnel-server.service /etc/systemd/system/
else
    # Create service file on the fly
    cat > /etc/systemd/system/tunnel-server.service << EOF
[Unit]
Description=GoTunnel Server - HTTP Tunnel Service
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=$ACTUAL_USER
Group=$ACTUAL_USER
ExecStart=/usr/local/bin/tunnel-server --port 8080 --domain easysource-mortalengine.hirequotient.com
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=tunnel-server

# Security settings
NoNewPrivileges=true
PrivateTmp=true

# Environment variables (optional)
Environment="LOG_LEVEL=info"

[Install]
WantedBy=multi-user.target
EOF
fi

chmod 644 /etc/systemd/system/tunnel-server.service
echo -e "${GREEN}✓ Service file created${NC}"

# 3. Enable and start service
echo -e "${BLUE}[3/4]${NC} Enabling and starting service..."
systemctl daemon-reload
systemctl enable tunnel-server.service
systemctl start tunnel-server.service
echo -e "${GREEN}✓ Service started${NC}"

# 4. Verify installation
echo -e "${BLUE}[4/4]${NC} Verifying installation..."
sleep 2
if systemctl is-active --quiet tunnel-server; then
    echo -e "${GREEN}✓ Service is running${NC}"
else
    echo -e "${RED}✗ Service failed to start${NC}"
    echo -e "${YELLOW}Check logs with: sudo journalctl -u tunnel-server -n 50${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Installation Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Show status
echo -e "${BLUE}Service Status:${NC}"
systemctl status tunnel-server.service --no-pager || true

echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo -e "  ${YELLOW}sudo systemctl status tunnel-server${NC}   # Check status"
echo -e "  ${YELLOW}sudo systemctl restart tunnel-server${NC}  # Restart"
echo -e "  ${YELLOW}sudo systemctl stop tunnel-server${NC}     # Stop"
echo -e "  ${YELLOW}sudo journalctl -u tunnel-server -f${NC}   # View logs"
echo ""
echo -e "${GREEN}Tunnel URL format: https://easysource-mortalengine.hirequotient.com/client-xxx/${NC}"
echo ""
