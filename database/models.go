package database

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Password will not be included in JSON
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpireAt  time.Time `json:"expire_at"`

	// Traffic limits
	TrafficLimit int64 `json:"traffic_limit"` // In bytes
	UsedTraffic  int64 `json:"used_traffic"`  // In bytes

	// User status
	Enabled bool `json:"enabled"`
	IsAdmin bool `json:"is_admin"`
}

// ProxyConfig represents a proxy configuration
type ProxyConfig struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Protocol  string    `json:"protocol"` // vmess, vless, trojan, shadowsocks
	Settings  string    `json:"settings"` // JSON string of protocol-specific settings
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Traffic statistics
	Upload   int64 `json:"upload"`   // In bytes
	Download int64 `json:"download"` // In bytes
}

// TrafficLog represents a traffic usage log entry
type TrafficLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProxyID   int64     `json:"proxy_id"`
	Upload    int64     `json:"upload"`   // In bytes
	Download  int64     `json:"download"` // In bytes
	Timestamp time.Time `json:"timestamp"`
}

// SystemStatus represents system status information
type SystemStatus struct {
	CPU     float64   `json:"cpu"`      // CPU usage percentage
	Memory  float64   `json:"memory"`   // Memory usage percentage
	Uptime  int64     `json:"uptime"`   // System uptime in seconds
	LoadAvg []float64 `json:"load_avg"` // System load averages
}
