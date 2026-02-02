package client

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"goTunnel/pkg/protocol"
)

// ForwardToLocal forwards an HTTP request to localhost and returns the response.
// This is the core function that makes goTunnel actually useful - it takes
// a tunneled HTTP request and forwards it to the local application.
//
// Key Go Concepts Demonstrated:
// - HTTP Client: Using http.Client to make requests
// - Error handling: Wrapping errors with context
// - Resource cleanup: Using defer to ensure response bodies are closed
func ForwardToLocal(reqData *protocol.HTTPRequestData, localPort string) (*protocol.HTTPResponseData, error) {
	// Build the target URL
	// We're forwarding to localhost on the specified port
	targetURL := fmt.Sprintf("http://localhost:%s", localPort)

	log.Printf("[DEBUG] Forwarding request to %s%s", targetURL, reqData.URL)

	// Convert the protocol request data to an http.Request
	// This uses the helper function from protocol/http.go
	req, err := reqData.ToHTTPRequest(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Create an HTTP client with a timeout
	// Timeouts are important to prevent hanging forever if the local app doesn't respond
	// This is a Go best practice for production code
	client := &http.Client{
		Timeout: 30 * time.Second, // 30 second timeout for local requests
	}

	// Make the actual HTTP request to localhost
	// This is where the magic happens - we're acting as a reverse proxy
	resp, err := client.Do(req)
	if err != nil {
		// Common errors:
		// - Connection refused: local app not running
		// - Timeout: local app too slow to respond
		// - Network errors: rare for localhost, but possible
		return nil, fmt.Errorf("failed to forward request to localhost: %w", err)
	}

	// Convert the HTTP response to our protocol format
	// This uses another helper from protocol/http.go
	// Note: HTTPResponseFromHTTP closes resp.Body for us
	respData, err := protocol.HTTPResponseFromHTTP(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTTP response: %w", err)
	}

	log.Printf("[DEBUG] Received response from localhost: status=%d, body_size=%d bytes",
		respData.StatusCode, len(respData.Body))

	return respData, nil
}
