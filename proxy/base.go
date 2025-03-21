package proxy

import (
	"fmt"
	"net"

	"v/common"
	"v/logger"
)

// BaseServer represents the base server implementation
type BaseServer struct {
	*common.ProxyServer
	Logger   *logger.Logger
	Listener net.Listener
	Running  bool
}

// NewBaseServer creates a new base server
func NewBaseServer(logger *logger.Logger, proxy *common.ProxyServer) *BaseServer {
	return &BaseServer{
		ProxyServer: proxy,
		Logger:      logger,
		Running:     false,
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
		return fmt.Errorf("failed to create listener: %v", err)
	}

	s.Listener = listener
	s.Running = true

	s.Logger.Info("Server started on port %d", s.Port)

	// Handle connections
	go s.handleConnections()

	return nil
}

// Stop stops the server
func (s *BaseServer) Stop() error {
	if !s.Running {
		return nil
	}

	if err := s.Listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %v", err)
	}

	s.Running = false
	s.Logger.Info("Server stopped")

	return nil
}

// GetPort returns the server port
func (s *BaseServer) GetPort() int {
	return s.Port
}

// GetProtocol returns the protocol type
func (s *BaseServer) GetProtocol() common.ProtocolType {
	return common.ProtocolType(s.Protocol)
}

// handleConnections handles incoming connections
func (s *BaseServer) handleConnections() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			}
			return
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a single connection
func (s *BaseServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// This method should be overridden by specific protocol implementations
	s.Logger.Error("handleConnection not implemented")
}
