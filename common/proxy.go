package common

import (
	"io"
	"time"
)

// ProxyConfig represents a proxy configuration
type ProxyConfig struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Type         string    `json:"type"`
	Port         int       `json:"port"`
	Settings     string    `json:"settings"`
	Enabled      bool      `json:"enabled"`
	Upload       int64     `json:"upload"`
	Download     int64     `json:"download"`
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProxyInstance represents a proxy server instance
type ProxyInstance struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	Type         string                 `json:"type"`
	Port         int                    `json:"port"`
	Settings     map[string]interface{} `json:"settings"`
	Server       ServerInterface        `json:"-"`
	Enabled      bool                   `json:"enabled"`
	Upload       int64                  `json:"upload"`
	Download     int64                  `json:"download"`
	LastActiveAt time.Time              `json:"last_active_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ServerInterface represents a proxy server interface
type ServerInterface interface {
	Start() error
	Stop() error
	HandleConnection(conn io.ReadWriteCloser) error
}
