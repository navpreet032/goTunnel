package protocol

import "encoding/json"

// MessageType represents the type of message being sent through the tunnel
type MessageType string

// Message type constants define the different types of messages that can be sent
// between the client and server through the WebSocket tunnel
const (
	// MessageTypeRegister is sent by client when first connecting to register itself
	MessageTypeRegister MessageType = "REGISTER"

	// MessageTypeHTTPRequest is sent by server to client with an HTTP request to forward
	MessageTypeHTTPRequest MessageType = "HTTP_REQ"

	// MessageTypeHTTPResponse is sent by client to server with the HTTP response
	MessageTypeHTTPResponse MessageType = "HTTP_RES"

	// MessageTypePing is sent to keep the connection alive
	MessageTypePing MessageType = "PING"

	// MessageTypePong is the response to a ping message
	MessageTypePong MessageType = "PONG"

	// MessageTypeError is sent when an error occurs
	MessageTypeError MessageType = "ERROR"
)

// Message is the core structure for all communication between client and server.
// It uses JSON encoding for simplicity and includes all necessary fields for
// routing and matching requests with responses.
type Message struct {
	// Type indicates what kind of message this is (REGISTER, HTTP_REQ, etc.)
	Type MessageType `json:"type"`

	// RequestID is a unique identifier used to match HTTP requests with their responses.
	// This is crucial for handling concurrent requests over a single WebSocket connection.
	RequestID string `json:"request_id,omitempty"`

	// ClientID is the unique identifier for the client tunnel.
	// Used by server to route requests to the correct client.
	ClientID string `json:"client_id,omitempty"`

	// Data contains the actual payload of the message.
	// The structure varies based on the message Type:
	// - REGISTER: RegistrationData
	// - HTTP_REQ: HTTPRequestData
	// - HTTP_RES: HTTPResponseData
	// - ERROR: ErrorData
	// We use json.RawMessage so we can unmarshal it based on Type later.
	Data json.RawMessage `json:"data,omitempty"`
}

// RegistrationData is sent by the client when registering with the server
type RegistrationData struct {
	// ClientID is optionally provided by client, otherwise server will generate one
	ClientID string `json:"client_id,omitempty"`
}

// RegistrationResponse is sent by the server after successful registration
type RegistrationResponse struct {
	// ClientID is the assigned client identifier
	ClientID string `json:"client_id"`

	// TunnelURL is the public URL where this tunnel can be accessed
	TunnelURL string `json:"tunnel_url"`

	// Message provides any additional information
	Message string `json:"message,omitempty"`
}

// ErrorData contains error information
type ErrorData struct {
	// Code is an optional error code
	Code string `json:"code,omitempty"`

	// Message describes the error
	Message string `json:"message"`
}

// Marshal converts a Message to JSON bytes.
// This is a helper function that wraps json.Marshal for convenience.
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal converts JSON bytes into a Message.
// This is a helper function that wraps json.Unmarshal for convenience.
func Unmarshal(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewMessage creates a new Message with the specified type.
// This is a convenience function to ensure messages are properly initialized.
func NewMessage(msgType MessageType) *Message {
	return &Message{
		Type: msgType,
	}
}

// NewRegisterMessage creates a REGISTER message
func NewRegisterMessage(clientID string) (*Message, error) {
	regData := RegistrationData{
		ClientID: clientID,
	}
	data, err := json.Marshal(regData)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: MessageTypeRegister,
		Data: data,
	}, nil
}

// NewPingMessage creates a PING message
func NewPingMessage() *Message {
	return &Message{
		Type: MessageTypePing,
	}
}

// NewPongMessage creates a PONG message
func NewPongMessage() *Message {
	return &Message{
		Type: MessageTypePong,
	}
}

// NewErrorMessage creates an ERROR message
func NewErrorMessage(code, message string) (*Message, error) {
	errData := ErrorData{
		Code:    code,
		Message: message,
	}
	data, err := json.Marshal(errData)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type: MessageTypeError,
		Data: data,
	}, nil
}
