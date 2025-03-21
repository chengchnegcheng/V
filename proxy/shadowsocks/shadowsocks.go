package shadowsocks

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"v/common"
	"v/logger"
	"v/proxy"
)

// Config represents Shadowsocks configuration
type Config struct {
	Method    string `json:"method"`
	Password  string `json:"password"`
	Obfs      string `json:"obfs,omitempty"`
	ObfsParam string `json:"obfs_param,omitempty"`
}

// Server represents a Shadowsocks server
type Server struct {
	logger    *logger.Logger
	proxy     *common.ProxyServer
	config    *common.ShadowsocksConfig
	tlsConfig *tls.Config
	block     cipher.Block
	listener  net.Listener
}

// New creates a new Shadowsocks server
func New(logger *logger.Logger, proxy *common.ProxyServer) (*Server, error) {
	// Parse config
	var config common.ShadowsocksConfig
	if err := json.Unmarshal([]byte(proxy.Settings["shadowsocks"].(string)), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	server := &Server{
		logger: logger,
		proxy:  proxy,
		config: &config,
	}

	// Setup cipher
	key := sha256.Sum256([]byte(config.Password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}
	server.block = block

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

// Start starts the server
func (s *Server) Start() error {
	var err error
	addr := fmt.Sprintf(":%d", s.proxy.Port)

	// Create listener
	if s.tlsConfig != nil {
		s.listener, err = tls.Listen("tcp", addr, s.tlsConfig)
	} else {
		s.listener, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}

	s.logger.Info("Shadowsocks server started on port %d", s.proxy.Port)

	// Handle connections
	go s.accept()

	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
		}
		s.listener = nil
	}

	s.logger.Info("Shadowsocks server stopped")
	return nil
}

// GetPort gets the server port
func (s *Server) GetPort() int {
	return s.proxy.Port
}

// GetProtocol gets the protocol type
func (s *Server) GetProtocol() common.ProtocolType {
	return common.ProtocolShadowsocks
}

// accept accepts connections
func (s *Server) accept() {
	for {
		conn, err := s.listener.Accept()
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
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read Shadowsocks header
	header, err := s.readHeader(conn)
	if err != nil {
		s.logger.Error("Failed to read header: %v", err)
		return
	}

	// Connect to target
	target, err := net.Dial("tcp", header.Address)
	if err != nil {
		s.logger.Error("Failed to connect to target: %v", err)
		return
	}
	defer target.Close()

	// Start proxying
	go proxy.CopyIO(conn, target)
	proxy.CopyIO(target, conn)
}

// Header represents a Shadowsocks header
type Header struct {
	AddressType byte
	Address     string
	Port        uint16
	Command     byte
}

// readHeader reads and parses the Shadowsocks header
func (s *Server) readHeader(conn net.Conn) (*Header, error) {
	// Read address type
	addrType := make([]byte, 1)
	if _, err := io.ReadFull(conn, addrType); err != nil {
		return nil, fmt.Errorf("failed to read address type: %v", err)
	}

	// Read address
	var addr string
	var port uint16
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

	// Read port
	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return nil, fmt.Errorf("failed to read port: %v", err)
	}
	port = uint16(portBytes[0])<<8 | uint16(portBytes[1])

	return &Header{
		AddressType: addrType[0],
		Address:     fmt.Sprintf("%s:%d", addr, port),
		Port:        port,
		Command:     1, // TCP
	}, nil
}
