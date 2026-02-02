# Deployment Guide

## Quick Setup for EC2 Server

### 1. Create GitHub Repository

```bash
# On your local machine (already done)
cd /Users/admin/Desktop/goTunnel

# Create a new repository on GitHub, then:
git remote add origin https://github.com/navpreet032/goTunnel.git
git branch -M main
git push -u origin main
```

### 2. On Your EC2 Server

```bash
# Install Go (if not already installed)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Clone the repository
git clone https://github.com/navpreet032/goTunnel.git
cd goTunnel

# Build the server
go build -o bin/server ./cmd/server

# Start the server
./scripts/start-server.sh 8080

# Or with custom domain
./scripts/start-server.sh 8080 your-domain.com
```

### 3. On Your Local Machine

```bash
# Navigate to project directory
cd /Users/admin/Desktop/goTunnel

# Start your local web app (example: Vite dev server)
npm run dev  # or whatever starts your app on port 3000/5173

# In another terminal, start the tunnel client
./scripts/start-client.sh ws://YOUR-EC2-IP:8080/tunnel 3000

# Example with actual IP
./scripts/start-client.sh ws://54.123.45.67:8080/tunnel 5173
```

### 4. Test the Tunnel

```bash
# From anywhere (your laptop, phone, etc.)
curl http://YOUR-EC2-IP:8080/client-1/

# Or open in browser
http://YOUR-EC2-IP:8080/client-1/
```

## Systemd Service (Optional - for production)

To run the server as a background service:

```bash
# Create systemd service file
sudo nano /etc/systemd/system/gotunnel.service
```

Paste this content:

```ini
[Unit]
Description=GoTunnel Server
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/goTunnel
ExecStart=/home/ubuntu/goTunnel/bin/server --port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Then:

```bash
# Enable and start the service
sudo systemctl enable gotunnel
sudo systemctl start gotunnel

# Check status
sudo systemctl status gotunnel

# View logs
sudo journalctl -u gotunnel -f
```

## Firewall Configuration

Make sure port 8080 is open on your EC2 security group:

- **Type**: Custom TCP
- **Port Range**: 8080
- **Source**: 0.0.0.0/0 (allow from anywhere)

## Troubleshooting

### Server not accessible

```bash
# Check if server is running
ps aux | grep server

# Check if port is listening
sudo netstat -tlnp | grep 8080

# Check firewall
sudo ufw status
```

### WebSocket connection fails

- Ensure you're using `ws://` (not `wss://` for now)
- Verify EC2 security group allows inbound on port 8080
- Check server logs for errors

## Using with HTTPS (Future Phase)

For production with HTTPS:

1. Set up a domain name pointing to your EC2 IP
2. Install SSL certificate (Let's Encrypt)
3. Use reverse proxy (nginx) or update server to support TLS
4. Change client connection to `wss://` (WebSocket Secure)
