package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	"v/common"
	"v/logger"
)

// VLESSServer represents a VLESS server
type VLESSServer struct {
	*BaseServer
	config    *common.VLESSConfig
	tlsConfig *tls.Config
}

// NewVLESSServer creates a new VLESS server
func NewVLESSServer(logger *logger.Logger, config *common.VLESSConfig, proxy *common.ProxyServer) (*VLESSServer, error) {
	server := &VLESSServer{
		BaseServer: NewBaseServer(logger, proxy),
		config:     config,
	}

	// Setup TLS if enabled
	if config.Security == "tls" {
		proxyConfig := proxy.Settings["tls"].(map[string]interface{})
		if proxyConfig != nil {
			cert, err := tls.LoadX509KeyPair(
				proxyConfig["cert_file"].(string),
				proxyConfig["key_file"].(string),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS certificate: %v", err)
			}

			server.tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				ServerName:   proxyConfig["server_name"].(string),
			}
		}
	}

	return server, nil
}

// Start starts the VLESS server
func (s *VLESSServer) Start() error {
	if s.Running {
		return fmt.Errorf("server is already running")
	}

	var err error
	addr := fmt.Sprintf(":%d", s.Port)

	// Create listener
	if s.tlsConfig != nil {
		s.Listener, err = tls.Listen("tcp", addr, s.tlsConfig)
	} else {
		s.Listener, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}

	s.Running = true
	s.Logger.Info("VLESS server started on port %d", s.Port)

	// Handle connections
	go s.handleConnections()

	return nil
}

// Stop stops the VLESS server
func (s *VLESSServer) Stop() error {
	if !s.Running {
		return nil
	}

	if err := s.Listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %v", err)
	}

	s.Running = false
	s.Logger.Info("VLESS server stopped")
	return nil
}

// handleConnections handles incoming connections
func (s *VLESSServer) handleConnections() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(time.Second)
				continue
			}
			return
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a single connection
func (s *VLESSServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read VLESS header
	header, err := s.readHeader(conn)
	if err != nil {
		s.Logger.Error("Failed to read header: %v", err)
		return
	}

	// Verify user ID
	if header.ID != s.config.ID {
		s.Logger.Error("Invalid user ID: %s", header.ID)
		return
	}

	// Connect to target
	target, err := net.Dial("tcp", header.Address)
	if err != nil {
		s.Logger.Error("Failed to connect to target: %v", err)
		return
	}
	defer target.Close()

	// Start proxying
	go CopyIO(conn, target)
	CopyIO(target, conn)
}

// VLESSHeader represents a VLESS header
type VLESSHeader struct {
	Version     byte
	ID          string
	Command     byte
	Port        uint16
	Address     string
	AddressType byte
}

// readHeader reads and parses the VLESS header
func (s *VLESSServer) readHeader(conn net.Conn) (*VLESSHeader, error) {
	// Read version
	version := make([]byte, 1)
	if _, err := io.ReadFull(conn, version); err != nil {
		return nil, fmt.Errorf("failed to read version: %v", err)
	}
	if version[0] != 0 {
		return nil, fmt.Errorf("unsupported version: %d", version[0])
	}

	// Read user ID
	id := make([]byte, 16)
	if _, err := io.ReadFull(conn, id); err != nil {
		return nil, fmt.Errorf("failed to read user ID: %v", err)
	}

	// Read command
	command := make([]byte, 1)
	if _, err := io.ReadFull(conn, command); err != nil {
		return nil, fmt.Errorf("failed to read command: %v", err)
	}

	// Read port
	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return nil, fmt.Errorf("failed to read port: %v", err)
	}
	port := uint16(portBytes[0])<<8 | uint16(portBytes[1])

	// Read address type
	addrType := make([]byte, 1)
	if _, err := io.ReadFull(conn, addrType); err != nil {
		return nil, fmt.Errorf("failed to read address type: %v", err)
	}

	// Read address
	var addr string
	switch addrType[0] {
	case 1: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return nil, fmt.Errorf("failed to read IPv4 address: %v", err)
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
	case 2: // Domain
		lenByte := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenByte); err != nil {
			return nil, fmt.Errorf("failed to read domain length: %v", err)
		}
		domain := make([]byte, lenByte[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, fmt.Errorf("failed to read domain: %v", err)
		}
		addr = string(domain)
	case 3: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return nil, fmt.Errorf("failed to read IPv6 address: %v", err)
		}
		addr = fmt.Sprintf("[%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x]",
			ip[0], ip[1], ip[2], ip[3], ip[4], ip[5], ip[6], ip[7],
			ip[8], ip[9], ip[10], ip[11], ip[12], ip[13], ip[14], ip[15])
	default:
		return nil, fmt.Errorf("unsupported address type: %d", addrType[0])
	}

	return &VLESSHeader{
		Version:     version[0],
		ID:          string(id),
		Command:     command[0],
		Port:        port,
		Address:     fmt.Sprintf("%s:%d", addr, port),
		AddressType: addrType[0],
	}, nil
}
