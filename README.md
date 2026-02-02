# GoTunnel

A simple HTTP tunnel service built in Go, similar to ngrok or Cloudflare Tunnel. Expose your localhost to the internet using WebSocket tunneling.

## Architecture

```
Internet User → Server (EC2) → WebSocket Tunnel → Client → localhost:PORT
             (Public HTTP)     (Bidirectional)    (HTTP Proxy)
```

## Features

- ✅ WebSocket-based tunneling for HTTP requests
- ✅ Random unique client IDs (8-character alphanumeric)
- ✅ Multiple simultaneous clients support
- ✅ Smart routing based on client ID
- ✅ Bidirectional HTTP request/response forwarding
- ✅ Error handling (502 for local server down, 404 for missing clients)
- ✅ Keep-alive ping/pong mechanism
- ✅ Concurrent request handling

## Prerequisites

- Go 1.21 or higher
- A server with a public IP (e.g., EC2 instance)

## Installation

### Clone the repository

```bash
git clone <your-repo-url>
cd goTunnel
```

### Install dependencies

```bash
go mod download
```

### Build

```bash
# Build both server and client
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

## Usage

### On Your Server (EC2)

1. **Start the tunnel server**:

```bash
# Using the start script
./scripts/start-server.sh

# Or manually
./bin/server --port 8080
```

The server will listen on port 8080 for:
- WebSocket connections at `/tunnel`
- HTTP requests at `/<client-id>/`

### On Your Local Machine

1. **Start your local application** (e.g., a web server on port 3000):

```bash
# Example: Python HTTP server
python3 -m http.server 3000

# Or any other web application
npm run dev  # Next.js, Vite, etc.
```

2. **Start the tunnel client**:

```bash
# Using the start script
./scripts/start-client.sh <server-url> <local-port>

# Example
./scripts/start-client.sh ws://your-server.com:8080/tunnel 3000

# Or manually
./bin/client --server ws://your-server.com:8080/tunnel --local-port 3000
```

3. **Access your local app through the tunnel**:

The client will display your tunnel URL:
```
[INFO] Your tunnel URL: http://client-1:8080
```

Make requests to: `http://your-server.com:8080/client-1/`

## Configuration

### Server Options

```bash
./bin/server \
  --port 8080 \
  --domain tunnel.example.com  # Optional: custom domain
```

### Client Options

```bash
./bin/client \
  --server ws://localhost:8080/tunnel \
  --local-port 3000
```

## Development

### Project Structure

```
goTunnel/
├── cmd/
│   ├── server/          # Server entry point
│   └── client/          # Client entry point
├── pkg/
│   ├── protocol/        # Message types and HTTP helpers
│   ├── registry/        # Client connection registry
│   ├── server/          # Server HTTP handler
│   └── client/          # Client proxy logic
├── scripts/             # Start scripts
└── bin/                 # Compiled binaries
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Example: Testing with a Local Web App

1. **Start a simple HTTP server**:

```bash
python3 -m http.server 3000
```

2. **Start the tunnel server** (on your EC2):

```bash
./bin/server --port 8080
```

3. **Start the tunnel client** (on your local machine):

```bash
./bin/client --server ws://your-ec2-ip:8080/tunnel --local-port 3000
```

4. **Access from anywhere**:

```bash
curl http://your-ec2-ip:8080/client-1/
```

You should see the directory listing from your local Python server!

## Implementation Status

- ✅ **Phase 1**: Foundation - WebSocket connections
- ✅ **Phase 2**: HTTP Tunneling - Server to Client
- ✅ **Phase 3**: Local Forwarding - Client to Localhost
- ✅ **Phase 4**: Multiple Clients & Routing
- 📋 **Phase 5**: Reliability & Error Handling (Next)
- 📋 **Phase 6**: Polish & Documentation

## Troubleshooting

### Port already in use

```bash
# Kill process on port 8080 (macOS/Linux)
lsof -ti:8080 | xargs kill -9
```

### Connection refused

- Check if server is running and accessible
- Verify firewall rules allow traffic on server port
- Ensure WebSocket URL is correct

### 502 Bad Gateway

- Your local application is not running on the specified port
- Check local port number is correct

## License

MIT

## Contributing

Contributions welcome! This is an educational project to learn Go.
