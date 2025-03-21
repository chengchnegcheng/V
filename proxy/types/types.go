package types

import (
	"io"
)

// Server represents a proxy server interface
type Server interface {
	Start() error
	Stop() error
	HandleConnection(io.ReadWriteCloser) error
}

// Config represents the proxy configuration
type Config struct {
	Protocol string          `json:"protocol"`
	Settings interface{}     `json:"settings"`
	Stream   *StreamSettings `json:"stream,omitempty"`
}

// StreamSettings represents the stream settings
type StreamSettings struct {
	Network      string        `json:"network"`
	Security     string        `json:"security"`
	TLSSettings  *TLSSettings  `json:"tlsSettings,omitempty"`
	WSSettings   *WSSettings   `json:"wsSettings,omitempty"`
	HTTPSettings *HTTPSettings `json:"httpSettings,omitempty"`
}

// TLSSettings represents the TLS settings
type TLSSettings struct {
	ServerName string `json:"serverName"`
	CertFile   string `json:"certFile"`
	KeyFile    string `json:"keyFile"`
}

// WSSettings represents the WebSocket settings
type WSSettings struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers,omitempty"`
}

// HTTPSettings represents the HTTP settings
type HTTPSettings struct {
	Host    []string          `json:"host"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers,omitempty"`
}
