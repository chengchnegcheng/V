package proxy

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"v/common"
	"v/logger"
)

// TrojanConfig Trojan 配置
type TrojanConfig struct {
	Password string `json:"password"`
	SSL      struct {
		CertFile string `json:"cert_file"`
		KeyFile  string `json:"key_file"`
	} `json:"ssl"`
}

// TrojanServer Trojan 服务器
type TrojanServer struct {
	logger    *logger.Logger
	proxy     *common.ProxyServer
	config    *common.TrojanConfig
	listener  net.Listener
	tlsConfig *tls.Config
	mu        sync.Mutex
}

// NewTrojanServer 创建 Trojan 服务器
func NewTrojanServer(logger *logger.Logger, proxy *common.ProxyServer) (*TrojanServer, error) {
	// 解析配置
	var config common.TrojanConfig
	if err := json.Unmarshal([]byte(proxy.Settings["trojan"].(string)), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	server := &TrojanServer{
		logger: logger,
		proxy:  proxy,
		config: &config,
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

// Start 启动服务器
func (s *TrojanServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	s.logger.Info("Trojan server started on port %d", s.proxy.Port)

	// Handle connections
	go s.accept()

	return nil
}

// Stop 停止服务器
func (s *TrojanServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
		}
		s.listener = nil
	}

	s.logger.Info("Trojan server stopped")
	return nil
}

// GetPort 获取服务器端口
func (s *TrojanServer) GetPort() int {
	return s.proxy.Port
}

// GetProtocol 获取协议类型
func (s *TrojanServer) GetProtocol() common.ProtocolType {
	return common.ProtocolTrojan
}

// accept 接受连接
func (s *TrojanServer) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			}
			return
		}

		go s.handleConnection(conn)
	}
}

// handleConnection 处理连接
func (s *TrojanServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read Trojan header
	header, err := s.readHeader(conn)
	if err != nil {
		s.logger.Error("Failed to read header: %v", err)
		return
	}

	// Verify password hash
	if !s.verifyPassword(header.PasswordHash) {
		s.logger.Error("Invalid password hash")
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
	go CopyIO(conn, target)
	CopyIO(target, conn)
}

// TrojanHeader represents a Trojan header
type TrojanHeader struct {
	PasswordHash []byte
	Command      byte
	Address      string
	Port         uint16
}

// readHeader reads and parses the Trojan header
func (s *TrojanServer) readHeader(conn net.Conn) (*TrojanHeader, error) {
	// Read password hash
	hash := make([]byte, 56)
	if _, err := io.ReadFull(conn, hash); err != nil {
		return nil, fmt.Errorf("failed to read password hash: %v", err)
	}

	// Read command
	command := make([]byte, 1)
	if _, err := io.ReadFull(conn, command); err != nil {
		return nil, fmt.Errorf("failed to read command: %v", err)
	}

	// Read CRLF
	crlf := make([]byte, 2)
	if _, err := io.ReadFull(conn, crlf); err != nil {
		return nil, fmt.Errorf("failed to read CRLF: %v", err)
	}
	if crlf[0] != '\r' || crlf[1] != '\n' {
		return nil, fmt.Errorf("invalid CRLF")
	}

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

	return &TrojanHeader{
		PasswordHash: hash,
		Command:      command[0],
		Address:      fmt.Sprintf("%s:%d", addr, port),
		Port:         port,
	}, nil
}

// verifyPassword verifies the password hash
func (s *TrojanServer) verifyPassword(hash []byte) bool {
	// Calculate SHA256 hash of the password
	h := sha256.New()
	h.Write([]byte(s.config.Password))
	expectedHash := h.Sum(nil)

	// Compare hashes
	for i := 0; i < 32; i++ {
		if hash[i] != expectedHash[i] {
			return false
		}
	}
	return true
}
