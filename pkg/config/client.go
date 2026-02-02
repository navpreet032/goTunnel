package config

import (
	"flag"
	"os"
	"strconv"

	"goTunnel/pkg/logger"
)

// ClientConfig holds all configuration for the tunnel client
// This demonstrates Go's configuration management with env vars and flags
type ClientConfig struct {
	// ServerURL is the WebSocket URL of the tunnel server
	ServerURL string

	// LocalPort is the port of the local application to forward to
	LocalPort string

	// MaxReconnectDelay is the maximum delay between reconnection attempts (seconds)
	MaxReconnectDelay int

	// LogLevel controls the verbosity of logging
	LogLevel logger.LogLevel
}

// LoadClientConfig loads configuration from environment variables and CLI flags
// CLI flags take precedence over environment variables
func LoadClientConfig() *ClientConfig {
	cfg := &ClientConfig{}

	// Define flags
	serverURL := flag.String("server", getEnv("TUNNEL_SERVER", "ws://localhost:8080/tunnel"), "Tunnel server WebSocket URL (env: TUNNEL_SERVER)")
	localPort := flag.String("local-port", getEnv("LOCAL_PORT", "3000"), "Local port to forward requests to (env: LOCAL_PORT)")
	maxReconnectDelay := flag.Int("max-reconnect-delay", getEnvInt("MAX_RECONNECT_DELAY", 60), "Maximum reconnect delay in seconds (env: MAX_RECONNECT_DELAY)")
	logLevel := flag.String("log-level", getEnv("LOG_LEVEL", "info"), "Log level: debug, info, warn, error (env: LOG_LEVEL)")

	flag.Parse()

	// Populate config
	cfg.ServerURL = *serverURL
	cfg.LocalPort = *localPort
	cfg.MaxReconnectDelay = *maxReconnectDelay
	cfg.LogLevel = logger.LogLevel(*logLevel)

	return cfg
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
