package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPRequestData contains all the data needed to reconstruct an HTTP request.
// When the server receives an HTTP request from the internet, it converts it
// to this structure and sends it to the client through the WebSocket tunnel.
type HTTPRequestData struct {
	// Method is the HTTP method (GET, POST, PUT, DELETE, etc.)
	Method string `json:"method"`

	// URL is the full request URL (path + query string)
	URL string `json:"url"`

	// Headers contains all HTTP headers from the original request
	Headers map[string][]string `json:"headers"`

	// Body contains the raw request body bytes
	Body []byte `json:"body,omitempty"`
}

// HTTPResponseData contains all the data from an HTTP response.
// After the client forwards the request to localhost and gets a response,
// it converts the response to this structure and sends it back to the server.
type HTTPResponseData struct {
	// StatusCode is the HTTP status code (200, 404, 500, etc.)
	StatusCode int `json:"status_code"`

	// Headers contains all HTTP headers from the response
	Headers map[string][]string `json:"headers"`

	// Body contains the raw response body bytes
	Body []byte `json:"body,omitempty"`
}

// HTTPRequestFromHTTP converts a standard http.Request to HTTPRequestData.
// This is used by the server to convert incoming HTTP requests into a format
// that can be sent through the tunnel.
func HTTPRequestFromHTTP(r *http.Request) (*HTTPRequestData, error) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	// Close the original body
	r.Body.Close()

	// Create the request data
	reqData := &HTTPRequestData{
		Method:  r.Method,
		URL:     r.URL.String(),
		Headers: r.Header,
		Body:    body,
	}

	return reqData, nil
}

// ToHTTPRequest converts HTTPRequestData back into an http.Request.
// This is used by the client to reconstruct the request before forwarding
// it to the local application.
func (h *HTTPRequestData) ToHTTPRequest(targetURL string) (*http.Request, error) {
	// Create a new request
	req, err := http.NewRequest(h.Method, targetURL+h.URL, bytes.NewReader(h.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	req.Header = h.Headers

	return req, nil
}

// HTTPResponseFromHTTP converts a standard http.Response to HTTPResponseData.
// This is used by the client after forwarding the request to localhost
// to package the response for sending back through the tunnel.
func HTTPResponseFromHTTP(resp *http.Response) (*HTTPResponseData, error) {
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	// Close the original body
	resp.Body.Close()

	// Create the response data
	respData := &HTTPResponseData{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}

	return respData, nil
}

// WriteToHTTPResponse writes HTTPResponseData to an http.ResponseWriter.
// This is used by the server to send the tunneled response back to the
// original HTTP client.
func (h *HTTPResponseData) WriteToHTTPResponse(w http.ResponseWriter) error {
	// Copy headers
	for key, values := range h.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code
	w.WriteHeader(h.StatusCode)

	// Write body
	if len(h.Body) > 0 {
		_, err := w.Write(h.Body)
		if err != nil {
			return fmt.Errorf("failed to write response body: %w", err)
		}
	}

	return nil
}

// NewHTTPRequestMessage creates a Message containing an HTTP request.
// This is a helper function used by the server.
func NewHTTPRequestMessage(requestID, clientID string, reqData *HTTPRequestData) (*Message, error) {
	data, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HTTP request data: %w", err)
	}

	return &Message{
		Type:      MessageTypeHTTPRequest,
		RequestID: requestID,
		ClientID:  clientID,
		Data:      data,
	}, nil
}

// NewHTTPResponseMessage creates a Message containing an HTTP response.
// This is a helper function used by the client.
func NewHTTPResponseMessage(requestID, clientID string, respData *HTTPResponseData) (*Message, error) {
	data, err := json.Marshal(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal HTTP response data: %w", err)
	}

	return &Message{
		Type:      MessageTypeHTTPResponse,
		RequestID: requestID,
		ClientID:  clientID,
		Data:      data,
	}, nil
}
