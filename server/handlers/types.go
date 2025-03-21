package handlers

import (
	"time"
)

// UserResponse represents the user information in API responses
type UserResponse struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	IsAdmin      bool       `json:"is_admin"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpireAt     *time.Time `json:"expire_at,omitempty"`
	TrafficLimit int64      `json:"traffic_limit"`
	UsedTraffic  int64      `json:"used_traffic"`
}

// UpdateUserRequest represents a request to update user information
type UpdateUserRequest struct {
	Email        *string    `json:"email"`
	Password     *string    `json:"password"`
	IsAdmin      *bool      `json:"is_admin"`
	ExpireAt     *time.Time `json:"expire_at"`
	TrafficLimit *int64     `json:"traffic_limit"`
}

// UpdatePasswordRequest represents a request to update user password
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

// TrafficResponse represents traffic statistics in API responses
type TrafficResponse struct {
	TrafficLimit int64 `json:"traffic_limit"`
	UsedTraffic  int64 `json:"used_traffic"`
	Remaining    int64 `json:"remaining"`
}

// ProtocolResponse represents protocol information in API responses
type ProtocolResponse struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Port         int       `json:"port"`
	Settings     string    `json:"settings"`
	Enable       bool      `json:"enable"`
	TrafficLimit int64     `json:"traffic_limit"`
	TrafficUsed  int64     `json:"traffic_used"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateProtocolRequest represents a request to create a new protocol
type CreateProtocolRequest struct {
	Type         string `json:"type" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Port         int    `json:"port" binding:"required"`
	Settings     string `json:"settings" binding:"required"`
	TrafficLimit int64  `json:"traffic_limit"`
}

// UpdateProtocolRequest represents a request to update protocol information
type UpdateProtocolRequest struct {
	Name         *string `json:"name"`
	Port         *int    `json:"port"`
	Settings     *string `json:"settings"`
	Enable       *bool   `json:"enable"`
	TrafficLimit *int64  `json:"traffic_limit"`
}

// SystemStatusResponse represents system status information in API responses
type SystemStatusResponse struct {
	CPUUsage             float64   `json:"cpu_usage"`
	CPUCount             int       `json:"cpu_count"`
	LoadAvg              []float64 `json:"load_avg"`
	MemoryTotal          uint64    `json:"memory_total"`
	MemoryUsed           uint64    `json:"memory_used"`
	MemoryFree           uint64    `json:"memory_free"`
	MemoryUsage          float64   `json:"memory_usage"`
	DiskTotal            uint64    `json:"disk_total"`
	DiskUsed             uint64    `json:"disk_used"`
	DiskFree             uint64    `json:"disk_free"`
	DiskUsage            float64   `json:"disk_usage"`
	NetworkBytesSent     uint64    `json:"network_bytes_sent"`
	NetworkBytesReceived uint64    `json:"network_bytes_received"`
	NetworkPacketsSent   uint64    `json:"network_packets_sent"`
	NetworkPacketsRecv   uint64    `json:"network_packets_recv"`
	Uptime               int64     `json:"uptime"`
	CreatedAt            time.Time `json:"created_at"`
}
