package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goTunnel/pkg/client"
	"goTunnel/pkg/protocol"

	"github.com/gorilla/websocket"
)

// Client represents the tunnel client that connects to the server
// and forwards requests to localhost
type Client struct {
	// serverURL is the WebSocket URL of the tunnel server
	serverURL string

	// localPort is the port of the local application to forward to
	localPort string

	// clientID is our unique identifier (assigned by server)
	clientID string

	// tunnelURL is the public URL assigned to us by the server
	tunnelURL string

	// conn is the WebSocket connection to the server
	conn *websocket.Conn

	// heartbeatStop is used to signal the heartbeat goroutine to stop
	heartbeatStop chan struct{}
}

// NewClient creates a new tunnel client instance
func NewClient(serverURL, localPort string) *Client {
	return &Client{
		serverURL: serverURL,
		localPort: localPort,
	}
}

// Connect establishes a connection to the tunnel server
func (c *Client) Connect() error {
	log.Printf("[INFO] Connecting to tunnel server at %s...", c.serverURL)

	// Dial the WebSocket server
	// This is how we establish the WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	log.Printf("[INFO] Connected to tunnel server successfully!")

	return nil
}

// Register sends a registration message to the server
func (c *Client) Register() error {
	log.Printf("[INFO] Registering with tunnel server...")

	// Create a registration message
	// For now, we'll let the server assign us an ID
	regMsg, err := protocol.NewRegisterMessage("")
	if err != nil {
		return fmt.Errorf("failed to create registration message: %w", err)
	}

	// Send the registration message
	if err := c.conn.WriteJSON(regMsg); err != nil {
		return fmt.Errorf("failed to send registration message: %w", err)
	}

	// Wait for registration response
	// This demonstrates reading a specific message and parsing it
	_, msgBytes, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read registration response: %w", err)
	}

	// Parse the response
	msg, err := protocol.Unmarshal(msgBytes)
	if err != nil {
		return fmt.Errorf("failed to parse registration response: %w", err)
	}

	if msg.Type != protocol.MessageTypeRegister {
		return fmt.Errorf("unexpected message type: %s", msg.Type)
	}

	// Parse the registration response data
	var regResp protocol.RegistrationResponse
	if err := json.Unmarshal(msg.Data, &regResp); err != nil {
		return fmt.Errorf("failed to parse registration response data: %w", err)
	}

	// Store our assigned details
	c.clientID = regResp.ClientID
	c.tunnelURL = regResp.TunnelURL

	log.Printf("[INFO] ==========================================")
	log.Printf("[INFO] Registration successful!")
	log.Printf("[INFO] Client ID: %s", c.clientID)
	log.Printf("[INFO] Your tunnel URL: %s", c.tunnelURL)
	log.Printf("[INFO] Forwarding to: http://localhost:%s", c.localPort)
	log.Printf("[INFO] ==========================================")

	return nil
}

// Run starts the main message reading loop
// This goroutine continuously reads messages from the server
func (c *Client) Run() error {
	log.Printf("[INFO] Client is running. Press Ctrl+C to stop.")

	// Message reading loop
	// This demonstrates Go's simple approach to continuous processing
	for {
		// Read message from server
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			// Check if this is a normal closure
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ERROR] WebSocket error: %v", err)
			}
			return err
		}

		// Parse the message
		msg, err := protocol.Unmarshal(msgBytes)
		if err != nil {
			log.Printf("[ERROR] Failed to parse message: %v", err)
			continue
		}

		log.Printf("[INFO] Received message type: %s", msg.Type)

		// Handle different message types
		// In Phase 1, we only handle PING and HTTP_REQ (which we'll just log)
		switch msg.Type {
		case protocol.MessageTypePing:
			c.handlePing()

		case protocol.MessageTypePong:
			// Server acknowledged our ping
			log.Printf("[DEBUG] Received pong from server")

		case protocol.MessageTypeHTTPRequest:
			// Log HTTP request details for Phase 2 testing
			c.handleHTTPRequest(msg)

		default:
			log.Printf("[WARN] Unknown message type: %s", msg.Type)
		}
	}
}

// handlePing responds to a ping message with a pong
func (c *Client) handlePing() {
	pong := protocol.NewPongMessage()
	if err := c.conn.WriteJSON(pong); err != nil {
		log.Printf("[ERROR] Failed to send pong: %v", err)
	} else {
		log.Printf("[DEBUG] Sent pong response")
	}
}

// handleHTTPRequest processes an HTTP request from the server
// This is Phase 3: we now forward the request to localhost and send back the response
func (c *Client) handleHTTPRequest(msg *protocol.Message) {
	// Parse the HTTP request data
	var reqData protocol.HTTPRequestData
	if err := json.Unmarshal(msg.Data, &reqData); err != nil {
		log.Printf("[ERROR] Failed to parse HTTP request: %v", err)
		c.sendErrorResponse(msg.RequestID, "Failed to parse request")
		return
	}

	log.Printf("[INFO] ==========================================")
	log.Printf("[INFO] Received HTTP Request:")
	log.Printf("[INFO]   Request ID: %s", msg.RequestID)
	log.Printf("[INFO]   Method: %s", reqData.Method)
	log.Printf("[INFO]   URL: %s", reqData.URL)
	log.Printf("[INFO]   Headers: %d", len(reqData.Headers))
	log.Printf("[INFO]   Body Size: %d bytes", len(reqData.Body))
	log.Printf("[INFO] ==========================================")

	// Forward the request to localhost using our new proxy function
	log.Printf("[INFO] Forwarding to localhost:%s%s", c.localPort, reqData.URL)
	respData, err := client.ForwardToLocal(&reqData, c.localPort)
	if err != nil {
		log.Printf("[ERROR] Failed to forward request: %v", err)
		// Send an error response back to server
		c.sendErrorResponse(msg.RequestID, fmt.Sprintf("Failed to forward request: %v", err))
		return
	}

	log.Printf("[INFO] Response received from localhost: status=%d", respData.StatusCode)

	// Create the HTTP response message
	// We use the same RequestID so the server can match it to the original request
	respMsg, err := protocol.NewHTTPResponseMessage(msg.RequestID, c.clientID, respData)
	if err != nil {
		log.Printf("[ERROR] Failed to create response message: %v", err)
		return
	}

	// Send the response back through the WebSocket tunnel
	if err := c.conn.WriteJSON(respMsg); err != nil {
		log.Printf("[ERROR] Failed to send response: %v", err)
		return
	}

	log.Printf("[INFO] Response sent back through tunnel successfully")
	log.Printf("[INFO] ==========================================")
}

// Close closes the connection to the server and stops the heartbeat
func (c *Client) Close() error {
	// Stop heartbeat first
	c.stopHeartbeat()

	// Then close the connection
	if c.conn != nil {
		log.Printf("[INFO] Closing connection to server...")
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// sendErrorResponse sends an error response back to the server
// This is used when we fail to process a request locally
func (c *Client) sendErrorResponse(requestID, errorMsg string) {
	// Create a simple error response with 502 Bad Gateway status
	respData := &protocol.HTTPResponseData{
		StatusCode: http.StatusBadGateway,
		Headers:    make(map[string][]string),
		Body:       []byte(errorMsg),
	}

	// Set Content-Type header
	respData.Headers["Content-Type"] = []string{"text/plain"}

	// Create and send the response message
	respMsg, err := protocol.NewHTTPResponseMessage(requestID, c.clientID, respData)
	if err != nil {
		log.Printf("[ERROR] Failed to create error response message: %v", err)
		return
	}

	if err := c.conn.WriteJSON(respMsg); err != nil {
		log.Printf("[ERROR] Failed to send error response: %v", err)
		return
	}

	log.Printf("[INFO] Sent error response for request %s", requestID)
}

// startHeartbeat starts a goroutine that sends periodic pings
// This demonstrates Go's goroutines and time.Ticker with proper cleanup
func (c *Client) startHeartbeat() {
	// Create stop channel for this heartbeat
	c.heartbeatStop = make(chan struct{})

	// Create a ticker that fires every 30 seconds
	// This is a common Go pattern for periodic tasks
	ticker := time.NewTicker(30 * time.Second)

	// Start a goroutine (lightweight thread) to handle the heartbeat
	// This runs concurrently with the main message reading loop
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Send a ping message
				ping := protocol.NewPingMessage()
				if err := c.conn.WriteJSON(ping); err != nil {
					log.Printf("[ERROR] Failed to send ping: %v", err)
					return
				}
				log.Printf("[DEBUG] Sent ping")

			case <-c.heartbeatStop:
				// Stop signal received
				log.Printf("[DEBUG] Heartbeat stopped")
				return
			}
		}
	}()

	log.Printf("[INFO] Heartbeat started (30s interval)")
}

// stopHeartbeat stops the heartbeat goroutine
func (c *Client) stopHeartbeat() {
	if c.heartbeatStop != nil {
		close(c.heartbeatStop)
		c.heartbeatStop = nil
	}
}

func main() {
	// Parse command-line flags
	serverURL := flag.String("server", "ws://localhost:8080/tunnel", "Tunnel server WebSocket URL")
	localPort := flag.String("local-port", "3000", "Local port to forward requests to")
	maxReconnectDelay := flag.Int("max-reconnect-delay", 60, "Maximum reconnect delay in seconds")
	flag.Parse()

	// Create client instance
	client := NewClient(*serverURL, *localPort)

	// Set up graceful shutdown
	// This demonstrates Go's signal handling for clean exits
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Reconnection loop with exponential backoff
	// This demonstrates Go's error recovery and retry patterns
	reconnectDelay := 1 * time.Second
	maxDelay := time.Duration(*maxReconnectDelay) * time.Second
	attemptCount := 0

	for {
		attemptCount++

		// Try to connect and register
		if err := client.Connect(); err != nil {
			log.Printf("[ERROR] Connection failed (attempt %d): %v", attemptCount, err)

			// Check if we received a shutdown signal
			select {
			case <-sigChan:
				log.Printf("[INFO] Shutdown signal received during reconnect")
				return
			case <-time.After(reconnectDelay):
				// Exponential backoff: double the delay each time, up to max
				reconnectDelay *= 2
				if reconnectDelay > maxDelay {
					reconnectDelay = maxDelay
				}
				log.Printf("[INFO] Retrying connection in %v...", reconnectDelay)
				continue
			}
		}

		// Successfully connected, now register
		if err := client.Register(); err != nil {
			log.Printf("[ERROR] Registration failed (attempt %d): %v", attemptCount, err)
			client.Close()

			// Check for shutdown signal
			select {
			case <-sigChan:
				log.Printf("[INFO] Shutdown signal received during registration")
				return
			case <-time.After(reconnectDelay):
				reconnectDelay *= 2
				if reconnectDelay > maxDelay {
					reconnectDelay = maxDelay
				}
				log.Printf("[INFO] Retrying registration in %v...", reconnectDelay)
				continue
			}
		}

		// Reset reconnect delay after successful connection
		reconnectDelay = 1 * time.Second
		attemptCount = 0
		log.Printf("[INFO] Successfully connected and registered")

		// Start heartbeat
		client.startHeartbeat()

		// Run client in a goroutine so we can handle shutdown signals
		errChan := make(chan error, 1)
		go func() {
			errChan <- client.Run()
		}()

		// Wait for either an error or a shutdown signal
		// This demonstrates Go's select statement for waiting on multiple channels
		select {
		case err := <-errChan:
			if err != nil {
				log.Printf("[ERROR] Client disconnected: %v", err)
				client.Close()

				// Check if it's a shutdown signal or a real error
				select {
				case <-sigChan:
					log.Printf("[INFO] Shutdown signal received")
					return
				default:
					// Connection lost, will retry
					log.Printf("[WARN] Connection lost, attempting to reconnect...")
					time.Sleep(reconnectDelay)
					continue
				}
			}
		case <-sigChan:
			log.Printf("[INFO] Shutdown signal received")
			client.Close()
			return
		}
	}
}
