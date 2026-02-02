# GoTunnel - Tunnel Service Implementation Plan

## Overview
Create a tunnel service in Go (like ngrok/Cloudflare Tunnel) to expose localhost to the internet using your EC2 instance at `https://easysource-mortalengine.hirequotient.com/`.

## Architecture

```
Internet User → EC2 Server → WebSocket Tunnel → Local Client → localhost:PORT
              (Public HTTP)    (Bidirectional)   (HTTP Proxy)
```

### Components
1. **Server** (runs on EC2): Accepts client connections, routes HTTP requests through tunnels
2. **Client** (runs locally): Connects to server, forwards requests to localhost
3. **Protocol**: WebSocket-based message protocol for bidirectional communication

### Why WebSocket?
- Simpler than raw TCP (built-in message framing)
- HTTP-compatible (works through proxies/firewalls)
- Great for learning Go's concurrency patterns
- Standard library support with `gorilla/websocket`

## File Structure

```
goTunnel/
├── cmd/
│   ├── server/
│   │   └── main.go              # Server entry point
│   └── client/
│       └── main.go              # Client entry point
├── pkg/
│   ├── protocol/
│   │   ├── message.go           # Message types and constants
│   │   └── http.go              # HTTP request/response helpers
│   ├── registry/
│   │   └── registry.go          # Client connection registry (thread-safe)
│   ├── server/
│   │   ├── tunnel_server.go     # WebSocket tunnel server
│   │   ├── http_handler.go      # HTTP request handler
│   │   └── client_manager.go    # Client connection management
│   ├── client/
│   │   ├── tunnel_client.go     # WebSocket tunnel client
│   │   └── proxy.go             # Local proxy logic
│   └── config/
│       ├── server.go            # Server configuration
│       └── client.go            # Client configuration
├── go.mod
├── go.sum
└── README.md
```

## Protocol Design

### Message Types
```go
type Message struct {
    Type      string          // "REGISTER", "HTTP_REQ", "HTTP_RES", "PING", "PONG", "ERROR"
    RequestID string          // Unique ID for request/response matching
    ClientID  string          // Unique client identifier
    Data      json.RawMessage // Payload varies by type
}
```

### Key Flows
1. **Client Registration**: Client connects → sends REGISTER → server assigns ID → returns tunnel URL
2. **HTTP Request**: User requests URL → server routes to client → client forwards to localhost → response flows back
3. **Keep-Alive**: Ping/pong every 30 seconds to detect disconnections

## Implementation Phases

### Phase 1: Foundation (Basic Communication)
**Goal**: Establish WebSocket connection between client and server

**Tasks**:
1. Initialize Go module (`go.mod`) with `gorilla/websocket` dependency
2. Create `pkg/protocol/message.go`:
   - Define `Message` struct with JSON tags
   - Define message type constants
   - Implement Marshal/Unmarshal helpers
3. Create basic server (`cmd/server/main.go`):
   - HTTP server on port 8080
   - WebSocket upgrade endpoint `/tunnel`
   - Accept connections, store in map (clientID → WebSocket conn)
   - Log connections
4. Create basic client (`cmd/client/main.go`):
   - Connect to server's WebSocket endpoint
   - Send REGISTER message with random client ID
   - Keep connection alive
   - Handle disconnections

**Go Concepts**: HTTP server, WebSocket upgrade, basic concurrency

**Verification**: Run server and client, verify connection logs appear

---

### Phase 2: HTTP Tunneling (Server → Client)
**Goal**: Forward HTTP requests from server to client through WebSocket

**Tasks**:
1. Create `pkg/registry/registry.go`:
   - Thread-safe map: clientID → WebSocket connection
   - Methods: `Register()`, `Unregister()`, `Get()`
   - Use `sync.RWMutex` for concurrency safety
2. Create `pkg/server/http_handler.go`:
   - Handle incoming HTTP requests on `/*`
   - Extract client ID from Host header (subdomain or path)
   - Convert HTTP request to `Message` (type: HTTP_REQ)
   - Generate unique request ID (use `uuid`)
   - Send message through client's WebSocket
   - Create response channel, wait for response (with timeout)
3. Update server to use registry and HTTP handler
4. Create `pkg/protocol/http.go`:
   - `HTTPRequestData` struct (method, URL, headers, body)
   - `HTTPResponseData` struct (status code, headers, body)
   - Conversion helpers

**Go Concepts**: Mutexes (sync.RWMutex), channels, select with timeout, UUID generation

**Verification**: Send HTTP request to server, verify message reaches client (log it)

---

### Phase 3: Local Forwarding (Client → Localhost)
**Goal**: Client forwards requests to localhost and sends responses back

**Tasks**:
1. Create `pkg/client/proxy.go`:
   - Function `forwardToLocal(req *HTTPRequestData, localPort int) (*HTTPResponseData, error)`
   - Make actual HTTP request to `localhost:PORT`
   - Capture response (status, headers, body using `io.ReadAll`)
   - Return as `HTTPResponseData`
2. Update client (`cmd/client/main.go`):
   - Add message reading loop (separate goroutine)
   - When `HTTP_REQ` arrives: call `forwardToLocal()`, send `HTTP_RES` back
   - Handle errors (connection refused, timeout)
3. Update server to handle `HTTP_RES` messages:
   - Match response to pending request using request ID
   - Send response to waiting channel
   - Convert to HTTP response and send to original requester

**Go Concepts**: HTTP client, goroutines, channel-based request/response matching, error handling

**Verification**: End-to-end test - start local HTTP server on port 3000, make request through tunnel, verify response

---

### Phase 4: Multiple Clients & Routing
**Goal**: Support multiple simultaneous clients with unique URLs

**Tasks**:
1. Client ID assignment:
   - Server generates random client ID on registration (8-character alphanumeric)
   - Send back assigned tunnel URL (e.g., `http://abc12345.tunnel.com`)
2. Request routing:
   - Parse `Host` header from incoming HTTP requests
   - Extract client ID from subdomain (e.g., `abc12345.tunnel.com` → `abc12345`)
   - Look up client in registry
   - Forward to correct WebSocket connection
   - Handle "client not found" errors
3. Concurrent request handling:
   - Use goroutine for each incoming HTTP request
   - Separate goroutine for WebSocket message reading per client
   - Map of pending requests: requestID → response channel

**Go Concepts**: Goroutines per connection, concurrent map access with mutexes, HTTP Host header parsing

**Verification**: Run 2-3 clients simultaneously, send requests to each unique URL, verify correct routing

---

### Phase 5: Reliability & Error Handling
**Goal**: Make it robust and handle failures gracefully

**Tasks**:
1. Connection management:
   - Implement ping/pong heartbeat (30-second interval)
   - Client auto-reconnect logic (5-second backoff)
   - Graceful shutdown on SIGINT/SIGTERM (use `signal.Notify`)
2. Error handling:
   - Client disconnection cleanup (remove from registry)
   - Request timeout handling (5 seconds default)
   - Invalid message handling (log and skip)
   - Local server unreachable (return 502 Bad Gateway)
3. Configuration:
   - Client: server URL (env var `TUNNEL_SERVER`), local port (`LOCAL_PORT`)
   - Server: listen port (`SERVER_PORT`), domain (`TUNNEL_DOMAIN`)
   - Use `flag` package for CLI arguments
4. Logging:
   - Use `log/slog` for structured logging
   - Log levels: INFO, WARN, ERROR, DEBUG
   - Log all connections, disconnections, requests, errors

**Go Concepts**: Context cancellation, signal handling, defer cleanup, structured logging, flag parsing

**Verification**: Kill client, verify reconnection; kill local server, verify 502; send invalid messages, verify handling

---

### Phase 6: Polish & Documentation
**Goal**: Make it user-friendly and educational

**Tasks**:
1. CLI improvements:
   - Client: Pretty output with assigned tunnel URL, request logs
   - Server: Startup message with endpoints, active connection count
   - Version flag (`--version`), help text (`--help`)
2. README documentation:
   - Architecture diagram (ASCII art)
   - Prerequisites (Go version, dependencies)
   - Installation instructions (`go install`)
   - Usage examples (server and client)
   - How it works section (educational)
   - Troubleshooting common issues
3. Code comments:
   - Document key functions and structs
   - Explain Go concepts in comments (for learning)
   - Add examples in docstrings

**Go Concepts**: CLI design, documentation best practices

**Verification**: Follow README from scratch, verify everything works; test with a real local app (e.g., React dev server)

---

## Critical Files

After implementation, these will be the core files:

1. **[cmd/server/main.go](cmd/server/main.go)** - Server entry point, orchestrates server components
2. **[pkg/server/http_handler.go](pkg/server/http_handler.go)** - HTTP request handling and tunneling logic
3. **[pkg/registry/registry.go](pkg/registry/registry.go)** - Thread-safe client connection management
4. **[cmd/client/main.go](cmd/client/main.go)** - Client entry point, manages tunnel and forwarding
5. **[pkg/protocol/message.go](pkg/protocol/message.go)** - Communication protocol, shared by both sides

## Key Go Concepts Used

- **Goroutines**: Concurrent request handling, message reading loops
- **Channels**: Request/response matching, timeouts
- **Mutexes**: Thread-safe registry (sync.RWMutex)
- **HTTP Server**: net/http package, custom handlers
- **WebSocket**: gorilla/websocket for bidirectional communication
- **Context**: Timeouts, cancellation propagation
- **Error Handling**: Explicit errors, wrapping with `fmt.Errorf`
- **JSON**: Encoding/decoding protocol messages
- **Defer**: Resource cleanup (connections, files)
- **Interfaces**: Abstraction for testing (optional)

## Verification Plan

### Unit Tests (Optional)
- Test message serialization/deserialization
- Test registry operations (concurrent access)
- Test HTTP data conversion helpers

### Integration Tests
1. **Basic Connection**: Client can connect to server
2. **HTTP Tunneling**: Request flows through tunnel to localhost
3. **Multiple Clients**: Requests route to correct clients
4. **Reconnection**: Client recovers from disconnection
5. **Timeouts**: Request timeout returns 504
6. **Errors**: Local server down returns 502

### Manual Testing Checklist
- [ ] Start server, verify startup message
- [ ] Start client, verify tunnel URL displayed
- [ ] Make HTTP request through tunnel, verify response
- [ ] Start second client, verify unique URLs
- [ ] Kill client, restart, verify reconnection
- [ ] Stop local server, verify 502 error
- [ ] Send large request, verify handling
- [ ] Graceful shutdown with Ctrl+C

## Example Usage

### Server
```bash
cd cmd/server
go run main.go --port 8080 --domain tunnel.example.com

# Output:
# [INFO] Tunnel server starting on :8080
# [INFO] Tunnel endpoint: ws://tunnel.example.com/tunnel
# [INFO] Ready to accept connections
```

### Client
```bash
cd cmd/client
go run main.go --server ws://localhost:8080/tunnel --local-port 3000

# Output:
# [INFO] Connecting to tunnel server...
# [INFO] Connected successfully!
# [INFO] Your tunnel URL: http://abc12345.localhost:8080
# [INFO] Forwarding to localhost:3000
# [INFO] Press Ctrl+C to stop
```

### Testing
```bash
# Start a test HTTP server
python3 -m http.server 3000

# Access via tunnel (in another terminal)
curl http://abc12345.localhost:8080

# Should see directory listing from Python HTTP server
```

## Educational Notes

This project teaches:
- **Networking fundamentals**: HTTP, WebSocket, TCP/IP
- **Go concurrency**: Goroutines, channels, synchronization
- **System design**: Client-server architecture, message protocols
- **Error handling**: Timeouts, reconnection, graceful degradation
- **Resource management**: Connection pooling, cleanup

Keep it simple - no over-engineering. Focus on understanding each Go concept as you implement it.

## Next Steps

After basic tunnel works, consider adding:
- TLS/HTTPS support (WSS protocol)
- Custom subdomain requests
- Authentication tokens
- Web dashboard showing active tunnels
- Request/response logging UI
- TCP tunneling (not just HTTP)
- Multiple local ports per client

## Simplifications (For Learning)

This implementation deliberately skips:
- Authentication (real tunnels use tokens)
- Rate limiting (could be abused)
- TLS encryption (using plain HTTP/WS)
- Input validation (trusting all data)
- Production-grade error recovery

These are important for production but add complexity that distracts from learning Go fundamentals.
