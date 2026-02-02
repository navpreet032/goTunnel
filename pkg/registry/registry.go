package registry

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Registry manages the mapping of client IDs to their WebSocket connections.
// This is a critical component that enables request routing: when an HTTP
// request comes in for a specific client, we use this registry to find
// the correct WebSocket connection to forward the request through.
//
// Key Go Concepts Demonstrated:
// - sync.RWMutex: Reader-writer mutex for efficient concurrent access
// - Thread-safe map operations
// - Defer for automatic lock release
type Registry struct {
	// clients maps clientID to WebSocket connection
	clients map[string]*websocket.Conn

	// mu protects the clients map from concurrent access
	// We use RWMutex (not just Mutex) because:
	// - Read operations (Get) are more frequent than writes
	// - Multiple readers can access simultaneously with RLock
	// - Only writes need exclusive Lock
	mu sync.RWMutex

	// pendingRequests maps requestID to a response channel
	// When we send an HTTP request through a tunnel, we create
	// a channel and wait for the response. This map allows us
	// to match incoming responses with waiting requests.
	pendingRequests map[string]chan interface{}

	// pendingMu protects the pendingRequests map
	pendingMu sync.RWMutex
}

// NewRegistry creates a new client registry instance
func NewRegistry() *Registry {
	return &Registry{
		clients:         make(map[string]*websocket.Conn),
		pendingRequests: make(map[string]chan interface{}),
	}
}

// Register adds a client connection to the registry
// This is called when a client successfully connects and registers
func (r *Registry) Register(clientID string, conn *websocket.Conn) error {
	// Acquire write lock - blocks all other reads and writes
	r.mu.Lock()
	// defer ensures Unlock happens when function exits
	// This is crucial: even if we panic or return early, the lock is released
	defer r.mu.Unlock()

	// Check if client ID already exists
	if _, exists := r.clients[clientID]; exists {
		return fmt.Errorf("client ID %s already registered", clientID)
	}

	r.clients[clientID] = conn
	return nil
}

// Unregister removes a client from the registry
// This is called when a client disconnects
func (r *Registry) Unregister(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from clients map
	delete(r.clients, clientID)
}

// Get retrieves a client connection by ID
// Returns the connection and a boolean indicating if it was found
//
// Note: We use RLock (not Lock) because this is a read-only operation
// Multiple goroutines can call Get simultaneously without blocking each other
func (r *Registry) Get(clientID string) (*websocket.Conn, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conn, ok := r.clients[clientID]
	return conn, ok
}

// GetAll returns a copy of all client IDs
// Useful for debugging and status reporting
func (r *Registry) GetAll() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a slice to hold all client IDs
	clientIDs := make([]string, 0, len(r.clients))
	for id := range r.clients {
		clientIDs = append(clientIDs, id)
	}

	return clientIDs
}

// Count returns the number of active clients
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.clients)
}

// RegisterPendingRequest stores a response channel for a request
// This creates the "waiting room" for a request's response
//
// Key Pattern: Request-Response Matching over Async Connection
// 1. Server sends HTTP request with unique requestID to client
// 2. Server creates a channel and stores it here
// 3. Server waits on the channel (with timeout)
// 4. Client processes request and sends response with same requestID
// 5. Server receives response, looks up channel by requestID, sends to channel
// 6. Waiting goroutine receives response and can continue
func (r *Registry) RegisterPendingRequest(requestID string, respChan chan interface{}) {
	r.pendingMu.Lock()
	defer r.pendingMu.Unlock()

	r.pendingRequests[requestID] = respChan
}

// UnregisterPendingRequest removes a pending request
// Called after response is received or request times out
func (r *Registry) UnregisterPendingRequest(requestID string) {
	r.pendingMu.Lock()
	defer r.pendingMu.Unlock()

	// Close and delete the channel
	if ch, exists := r.pendingRequests[requestID]; exists {
		close(ch)
		delete(r.pendingRequests, requestID)
	}
}

// GetPendingRequest retrieves a response channel for a request
// Returns nil if the request ID is not found
func (r *Registry) GetPendingRequest(requestID string) (chan interface{}, bool) {
	r.pendingMu.RLock()
	defer r.pendingMu.RUnlock()

	ch, ok := r.pendingRequests[requestID]
	return ch, ok
}

// PendingRequestCount returns the number of pending requests
// Useful for monitoring and debugging
func (r *Registry) PendingRequestCount() int {
	r.pendingMu.RLock()
	defer r.pendingMu.RUnlock()

	return len(r.pendingRequests)
}
