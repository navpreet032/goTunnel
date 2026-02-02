package config

import (
	"flag"

	"goTunnel/pkg/logger"
)

// ServerConfig holds all configuration for the tunnel server
type ServerConfig struct {
	// Port is the port the server listens on
	Port string

	// Domain is the domain name for tunnel URLs (optional)
	Domain string

	// RequestTimeout is how long to wait for client responses (seconds)
	RequestTimeout int

	// LogLevel controls the verbosity of logging
	LogLevel logger.LogLevel
}

// LoadServerConfig loads configuration from environment variables and CLI flags
// CLI flags take precedence over environment variables
func LoadServerConfig() *ServerConfig {
	cfg := &ServerConfig{}

	// Define flags
	port := flag.String("port", getEnv("SERVER_PORT", "8080"), "Port to listen on (env: SERVER_PORT)")
	domain := flag.String("domain", getEnv("TUNNEL_DOMAIN", ""), "Domain name for tunnel URLs (env: TUNNEL_DOMAIN)")
	requestTimeout := flag.Int("request-timeout", getEnvInt("REQUEST_TIMEOUT", 5), "Request timeout in seconds (env: REQUEST_TIMEOUT)")
	logLevel := flag.String("log-level", getEnv("LOG_LEVEL", "info"), "Log level: debug, info, warn, error (env: LOG_LEVEL)")

	flag.Parse()

	// Populate config
	cfg.Port = *port
	cfg.Domain = *domain
	cfg.RequestTimeout = *requestTimeout
	cfg.LogLevel = logger.LogLevel(*logLevel)

	return cfg
}
