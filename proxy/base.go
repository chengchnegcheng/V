package proxy

import (
	"fmt"
	"io"
	"net"
	"time"

	"v/common"
	"v/logger"
)

// BaseServer represents the base server implementation
type BaseServer struct {
	*common.ProxyInstance
	Logger       *logger.Logger
	Listener     net.Listener
	Running      bool
	Upload       int64
	Download     int64
	LastActiveAt time.Time
}

// NewBaseServer creates a new base server
func NewBaseServer(logger *logger.Logger, proxy *common.ProxyInstance) *BaseServer {
	return &BaseServer{
		ProxyInstance: proxy,
		Logger:        logger,
		Running:       false,
	}
}

// Start starts the server
func (s *BaseServer) Start() error {
	if s.Running {
		return fmt.Errorf("server is already running")
	}

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		s.Logger.Error("Failed to listen on port", logger.Fields{
			"port":  s.Port,
			"error": err.Error(),
		})
		return err
	}

	s.Listener = listener
	s.Running = true

	s.Logger.Info("Server started", logger.Fields{
		"port": s.Port,
	})

	// Handle connections
	go s.handleConnections()

	return nil
}

// Stop stops the server
func (s *BaseServer) Stop() error {
	if !s.Running {
		return fmt.Errorf("server is not running")
	}

	if err := s.Listener.Close(); err != nil {
		s.Logger.Error("Failed to close listener", logger.Fields{
			"error": err.Error(),
		})
		return err
	}

	s.Listener = nil
	s.Running = false

	s.Logger.Info("Server stopped", logger.Fields{
		"port": s.Port,
	})

	return nil
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

// GetProtocol returns the protocol type
func (s *BaseServer) GetProtocol() common.ProtocolType {
	return common.ProtocolType(s.Type)
}

// handleConnections handles incoming connections
func (s *BaseServer) handleConnections() {
	for s.Running {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.Running {
				s.Logger.Error("Failed to accept connection", logger.Fields{
					"error": err.Error(),
				})
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a connection
func (s *BaseServer) handleConnection(conn net.Conn) {
	// Base implementation does nothing
	defer conn.Close()
}

// HandleConnection implements the common.ProxyServerInterface interface
func (s *BaseServer) HandleConnection(conn io.ReadWriteCloser) error {
	// Base implementation does nothing
	return nil
}
