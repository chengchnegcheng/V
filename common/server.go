package common

import (
	"io"
	"net"
	"time"
)

// ProxyServerInterface represents a proxy server interface with extended functions
type ProxyServerInterface interface {
	Start() error
	Stop() error
	HandleConnection(conn io.ReadWriteCloser) error
	GetPort() int
	GetUpload() int64
	GetDownload() int64
	UpdateTraffic(upload, download int64)
	UpdateLastActive(time time.Time)
	GetLastActive() time.Time
}

// BaseServer implements basic server functionality
type BaseServer struct {
	Port         int
	Upload       int64
	Download     int64
	LastActiveAt time.Time
	Running      bool
	Listener     net.Listener
}

// GetPort returns the server port
func (s *BaseServer) GetPort() int {
	return s.Port
}

// GetUpload returns the upload traffic
func (s *BaseServer) GetUpload() int64 {
	return s.Upload
}

// GetDownload returns the download traffic
func (s *BaseServer) GetDownload() int64 {
	return s.Download
}

// UpdateTraffic updates traffic statistics
func (s *BaseServer) UpdateTraffic(upload, download int64) {
	s.Upload += upload
	s.Download += download
}

// UpdateLastActive updates last active time
func (s *BaseServer) UpdateLastActive(time time.Time) {
	s.LastActiveAt = time
}

// GetLastActive returns last active time
func (s *BaseServer) GetLastActive() time.Time {
	return s.LastActiveAt
}
