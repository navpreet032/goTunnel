package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"goTunnel/pkg/protocol"
	"goTunnel/pkg/registry"
	"goTunnel/pkg/server"

	"github.com/gorilla/websocket"
)

// Server represents the tunnel server that accepts client connections
// and manages the routing of HTTP requests through tunnels
type Server struct {
	// registry manages all client connections and pending requests
	registry *registry.Registry

	// upgrader is used to upgrade HTTP connections to WebSocket
	upgrader websocket.Upgrader

	// httpHandler handles HTTP requests and routes them through tunnels
	httpHandler *server.HTTPHandler

	// port is the port the server listens on
	port string

	// domain is the domain name for the tunnel service
	domain string
}

// NewServer creates a new tunnel server instance
func NewServer(port, domain string) *Server {
	// Create the registry to manage client connections
	reg := registry.NewRegistry()

	return &Server{
		registry: reg,
		upgrader: websocket.Upgrader{
			// CheckOrigin allows all connections for now (educational purposes)
			// In production, you'd want to validate the origin
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		httpHandler: server.NewHTTPHandler(reg),
		port:        port,
		domain:      domain,
	}
}

// handleTunnel handles WebSocket upgrade requests from clients wanting to establish a tunnel
func (s *Server) handleTunnel(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	// This is a key Go concept: protocol upgrade
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to upgrade connection: %v", err)
		return
	}

	log.Printf("[INFO] New client connection from %s", r.RemoteAddr)

	// Handle this client connection
	// We'll run this in the current goroutine for now (simple approach)
	s.handleClient(conn)
}

// handleClient manages a single client connection
// This function reads messages from the client and processes them
func (s *Server) handleClient(conn *websocket.Conn) {
	// Ensure we clean up when this function exits
	// defer is a Go feature that runs code when the function returns
	defer func() {
		conn.Close()
		log.Printf("[INFO] Client connection closed")
	}()

	var clientID string

	// Message reading loop
	// This continuously reads messages from the WebSocket connection
	for {
		// Read a message from the WebSocket
		// WebSocket handles the framing for us - much simpler than raw TCP!
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ERROR] WebSocket error: %v", err)
			}
			break
		}

		// Parse the message using our protocol
		msg, err := protocol.Unmarshal(msgBytes)
		if err != nil {
			log.Printf("[ERROR] Failed to unmarshal message: %v", err)
			continue
		}

		log.Printf("[INFO] Received message type: %s", msg.Type)

		// Handle different message types
		switch msg.Type {
		case protocol.MessageTypeRegister:
			clientID = s.handleRegister(conn, msg)

		case protocol.MessageTypePing:
			s.handlePing(conn, msg)

		case protocol.MessageTypeHTTPResponse:
			// Handle HTTP response from client
			if err := server.HandleHTTPResponse(s.registry, msg); err != nil {
				log.Printf("[ERROR] Failed to handle HTTP response: %v", err)
			}

		default:
			log.Printf("[WARN] Unknown message type: %s", msg.Type)
		}
	}

	// Clean up: remove client from registry when connection closes
	if clientID != "" {
		s.unregisterClient(clientID)
	}
}

// handleRegister processes a client registration request
func (s *Server) handleRegister(conn *websocket.Conn, msg *protocol.Message) string {
	// Parse the registration data
	var regData protocol.RegistrationData
	if err := json.Unmarshal(msg.Data, &regData); err != nil {
		log.Printf("[ERROR] Failed to parse registration data: %v", err)
		return ""
	}

	// Generate a client ID if not provided
	// For now, we'll use a simple counter-based approach
	// In later phases, we'll use UUIDs
	clientID := regData.ClientID
	if clientID == "" {
		clientID = s.generateClientID()
	}

	log.Printf("[INFO] Registering client: %s", clientID)

	// Store the client connection
	if err := s.registerClient(clientID, conn); err != nil {
		log.Printf("[ERROR] Failed to register client: %v", err)
		return ""
	}

	// Create the tunnel URL
	tunnelURL := fmt.Sprintf("http://%s:%s", clientID, s.port)
	if s.domain != "" {
		tunnelURL = fmt.Sprintf("http://%s.%s", clientID, s.domain)
	}

	// Send registration response back to client
	response := protocol.RegistrationResponse{
		ClientID:  clientID,
		TunnelURL: tunnelURL,
		Message:   "Registration successful",
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal registration response: %v", err)
		return clientID
	}

	responseMsg := &protocol.Message{
		Type: protocol.MessageTypeRegister,
		Data: responseData,
	}

	if err := conn.WriteJSON(responseMsg); err != nil {
		log.Printf("[ERROR] Failed to send registration response: %v", err)
		return clientID
	}

	log.Printf("[INFO] Client registered: %s -> %s", clientID, tunnelURL)

	return clientID
}

// handlePing responds to a ping message with a pong
func (s *Server) handlePing(conn *websocket.Conn, msg *protocol.Message) {
	pong := protocol.NewPongMessage()
	if err := conn.WriteJSON(pong); err != nil {
		log.Printf("[ERROR] Failed to send pong: %v", err)
	}
}

// registerClient adds a client to the registry
func (s *Server) registerClient(clientID string, conn *websocket.Conn) error {
	return s.registry.Register(clientID, conn)
}

// unregisterClient removes a client from the registry
func (s *Server) unregisterClient(clientID string) {
	s.registry.Unregister(clientID)
	log.Printf("[INFO] Unregistered client: %s", clientID)
}

// generateClientID creates a unique client identifier
// Simple implementation for Phase 2 - will be improved in Phase 4
func (s *Server) generateClientID() string {
	count := s.registry.Count()
	// For now, just use a counter
	// In Phase 4, we'll use proper random IDs
	return fmt.Sprintf("client-%d", count+1)
}

func main() {
	// Parse command-line flags
	// This demonstrates Go's flag package for CLI argument handling
	port := flag.String("port", "8080", "Port to listen on")
	domain := flag.String("domain", "", "Domain name for tunnel URLs (optional)")
	flag.Parse()

	// Create server instance
	srv := NewServer(*port, *domain)

	// Set up HTTP routes
	// The /tunnel endpoint is for WebSocket connections from clients
	http.HandleFunc("/tunnel", srv.handleTunnel)

	// All other requests are proxied through tunnels
	// This uses the HTTP handler to route requests to clients
	http.Handle("/", srv.httpHandler)

	// Start message
	listenAddr := fmt.Sprintf(":%s", *port)
	log.Printf("[INFO] ==========================================")
	log.Printf("[INFO] Tunnel server starting on %s", listenAddr)
	log.Printf("[INFO] Tunnel endpoint: ws://localhost:%s/tunnel", *port)
	log.Printf("[INFO] HTTP proxy endpoint: http://localhost:%s/<client-id>/", *port)
	log.Printf("[INFO] Ready to accept connections")
	log.Printf("[INFO] ==========================================")

	// Start the HTTP server
	// This will block and handle incoming connections
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}
