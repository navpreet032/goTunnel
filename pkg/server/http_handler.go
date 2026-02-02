package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"goTunnel/pkg/protocol"
	"goTunnel/pkg/registry"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// HTTPHandler handles incoming HTTP requests and routes them through tunnels
// This is the core of the tunneling functionality: it converts regular HTTP
// requests into tunnel messages, sends them to clients, and waits for responses
type HTTPHandler struct {
	// registry holds all active client connections
	registry *registry.Registry

	// requestTimeout is how long to wait for a response before giving up
	requestTimeout time.Duration
}

// NewHTTPHandler creates a new HTTP handler instance
func NewHTTPHandler(reg *registry.Registry) *HTTPHandler {
	return &HTTPHandler{
		registry:       reg,
		requestTimeout: 5 * time.Second, // 5 second default timeout
	}
}

// ServeHTTP implements http.Handler interface
// This is called for every HTTP request that comes to the server
// (except the /tunnel endpoint which is handled separately)
//
// Key Go Concept: http.Handler Interface
// Any type that implements ServeHTTP(ResponseWriter, *Request) can be
// used as an HTTP handler. This is Go's interface-based polymorphism.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract client ID from the request
	// We support two methods:
	// 1. Path-based: /clientID/path (e.g., /client-1/api/users)
	// 2. Subdomain-based: clientID.domain.com (for production use)
	clientID := h.extractClientID(r)

	if clientID == "" {
		http.Error(w, "Invalid request: no client ID found", http.StatusBadRequest)
		log.Printf("[WARN] No client ID found in request: %s %s", r.Method, r.URL.Path)
		return
	}

	log.Printf("[INFO] HTTP Request: %s %s -> client: %s", r.Method, r.URL.Path, clientID)

	// Look up the client connection in the registry
	conn, exists := h.registry.Get(clientID)
	if !exists {
		http.Error(w, fmt.Sprintf("Client %s not found or disconnected", clientID), http.StatusNotFound)
		log.Printf("[WARN] Client %s not found in registry", clientID)
		return
	}

	// Forward the request through the tunnel and wait for response
	if err := h.forwardRequest(w, r, clientID, conn); err != nil {
		log.Printf("[ERROR] Failed to forward request: %v", err)
		// If we haven't written a response yet, send error
		if w != nil {
			http.Error(w, "Failed to process request through tunnel", http.StatusBadGateway)
		}
	}
}

// extractClientID extracts the client ID from the request
// Supports both path-based and subdomain-based routing
func (h *HTTPHandler) extractClientID(r *http.Request) string {
	// Method 1: Path-based routing
	// Request format: http://server:8080/client-1/api/users
	// Extract "client-1" from the path
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 0 && parts[0] != "" && parts[0] != "tunnel" {
		return parts[0]
	}

	// Method 2: Subdomain-based routing (for future use)
	// Request format: http://client-1.tunnel.com/api/users
	// Extract "client-1" from Host header
	host := r.Host
	if strings.Contains(host, ".") {
		subdomain := strings.Split(host, ".")[0]
		// Remove port if present
		subdomain = strings.Split(subdomain, ":")[0]
		if subdomain != "" && subdomain != "localhost" {
			return subdomain
		}
	}

	return ""
}

// forwardRequest handles the complete request-response cycle through the tunnel
// This demonstrates several key Go patterns:
// - Channels for async communication
// - Select for timeout handling
// - Goroutine coordination
func (h *HTTPHandler) forwardRequest(w http.ResponseWriter, r *http.Request, clientID string, conn *websocket.Conn) error {
	// Generate a unique request ID
	// UUID ensures we can match responses even with concurrent requests
	requestID := uuid.New().String()

	// Convert the HTTP request to our protocol format
	reqData, err := protocol.HTTPRequestFromHTTP(r)
	if err != nil {
		return fmt.Errorf("failed to convert HTTP request: %w", err)
	}

	// Adjust the URL path if we extracted client ID from it
	// If request was /client-1/api/users, forward just /api/users
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) > 1 {
		reqData.URL = "/" + parts[1]
	}

	// Create the tunnel message
	msg, err := protocol.NewHTTPRequestMessage(requestID, clientID, reqData)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Create a response channel
	// This channel will receive the response from the client
	// Key Pattern: Using channels to wait for async responses
	respChan := make(chan interface{}, 1) // Buffered to avoid blocking sender

	// Register the pending request
	// This allows the message reader to find our channel when response arrives
	h.registry.RegisterPendingRequest(requestID, respChan)

	// Ensure cleanup: always unregister when we're done
	defer h.registry.UnregisterPendingRequest(requestID)

	// Send the request message to the client through WebSocket
	log.Printf("[DEBUG] Sending HTTP request to client %s (request ID: %s)", clientID, requestID)
	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("failed to send message to client: %w", err)
	}

	// Wait for response with timeout
	// This demonstrates Go's select statement for handling multiple channel operations
	select {
	case resp := <-respChan:
		// Got response! Convert and send back to client
		log.Printf("[DEBUG] Received response for request %s", requestID)
		return h.writeResponse(w, resp)

	case <-time.After(h.requestTimeout):
		// Timeout waiting for response
		log.Printf("[WARN] Request %s timed out after %v", requestID, h.requestTimeout)
		http.Error(w, "Gateway timeout: client did not respond in time", http.StatusGatewayTimeout)
		return nil
	}
}

// writeResponse writes the tunnel response back to the HTTP client
func (h *HTTPHandler) writeResponse(w http.ResponseWriter, resp interface{}) error {
	// Convert the response to HTTPResponseData
	respData, ok := resp.(*protocol.HTTPResponseData)
	if !ok {
		return fmt.Errorf("invalid response type")
	}

	// Write the response using our protocol helper
	if err := respData.WriteToHTTPResponse(w); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	log.Printf("[DEBUG] Response sent successfully (status: %d)", respData.StatusCode)
	return nil
}

// HandleHTTPResponse processes HTTP responses received from clients
// This is called when the server receives an HTTP_RES message from a client
// It matches the response to the waiting request and sends it through the channel
func HandleHTTPResponse(reg *registry.Registry, msg *protocol.Message) error {
	// Get the response channel for this request
	respChan, exists := reg.GetPendingRequest(msg.RequestID)
	if !exists {
		// Request might have timed out already
		log.Printf("[WARN] Received response for unknown request ID: %s", msg.RequestID)
		return fmt.Errorf("no pending request for ID: %s", msg.RequestID)
	}

	// Parse the HTTP response data
	var respData protocol.HTTPResponseData
	if err := json.Unmarshal(msg.Data, &respData); err != nil {
		return fmt.Errorf("failed to parse HTTP response: %w", err)
	}

	// Send the response to the waiting goroutine
	// This is the magic moment: the HTTP handler waiting on this channel
	// will receive the response and can complete the original HTTP request
	//
	// Key Pattern: Channel-based Request-Response Matching
	// Even though we're handling multiple concurrent requests,
	// each gets routed to the correct waiting handler via its unique channel
	select {
	case respChan <- &respData:
		log.Printf("[DEBUG] Response forwarded to pending request %s", msg.RequestID)
		return nil
	default:
		// Channel full or closed
		log.Printf("[WARN] Could not forward response for request %s", msg.RequestID)
		return fmt.Errorf("failed to forward response")
	}
}
