package proxy

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

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
	logger       *logger.Logger
	proxy        *common.ProxyInstance
	config       *common.TrojanConfig
	listener     net.Listener
	tlsConfig    *tls.Config
	mu           sync.Mutex
	port         int
	upload       int64
	download     int64
	lastActiveAt time.Time
	running      bool
}

// NewTrojanServer 创建 Trojan 服务器
func NewTrojanServer(logger *logger.Logger, proxy *common.ProxyInstance) (*TrojanServer, error) {
	// 解析配置
	var config common.TrojanConfig
	settingsStr, ok := proxy.Settings["trojan"]
	if !ok {
		return nil, fmt.Errorf("missing trojan settings")
	}

	// Handle different types of settings
	var settingsBytes []byte
	switch v := settingsStr.(type) {
	case string:
		settingsBytes = []byte(v)
	case map[string]interface{}:
		var err error
		settingsBytes, err = json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %v", err)
		}
	default:
		return nil, fmt.Errorf("invalid settings type: %T", settingsStr)
	}

	if err := json.Unmarshal(settingsBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	server := &TrojanServer{
		logger:       logger,
		proxy:        proxy,
		config:       &config,
		port:         proxy.Port,
		upload:       proxy.Upload,
		download:     proxy.Download,
		lastActiveAt: proxy.LastActiveAt,
	}

	// Setup TLS if enabled
	if config.Security == "tls" {
		proxyConfig, ok := proxy.Settings["tls"].(map[string]interface{})
		if !ok || proxyConfig == nil {
			return nil, fmt.Errorf("invalid TLS configuration")
		}

		certFile, ok1 := proxyConfig["cert_file"].(string)
		keyFile, ok2 := proxyConfig["key_file"].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("missing certificate files in TLS configuration")
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificate: %v", err)
		}

		serverName, ok := proxyConfig["server_name"].(string)
		if !ok {
			serverName = ""
		}

		server.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ServerName:   serverName,
		}
	}

	return server, nil
}

// Start 启动服务器
func (s *TrojanServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	var err error
	addr := fmt.Sprintf(":%d", s.port)

	// Create listener
	if s.tlsConfig != nil {
		s.listener, err = tls.Listen("tcp", addr, s.tlsConfig)
	} else {
		s.listener, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}

	s.running = true
	s.logger.Info("Trojan server started", logger.Fields{
		"port": s.port,
	})

	// Handle connections
	go s.accept()

	return nil
}

// Stop 停止服务器
func (s *TrojanServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("server is not running")
	}

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Error("Failed to close listener", logger.Fields{
				"error": err.Error(),
			})
			return err
		}
		s.listener = nil
	}

	s.running = false
	s.logger.Info("Trojan server stopped", logger.Fields{
		"port": s.port,
	})

	return nil
}

// GetPort returns the server port
func (s *TrojanServer) GetPort() int {
	return s.port
}

// GetUpload returns the upload traffic
func (s *TrojanServer) GetUpload() int64 {
	return s.upload
}

// GetDownload returns the download traffic
func (s *TrojanServer) GetDownload() int64 {
	return s.download
}

// UpdateTraffic updates traffic statistics
func (s *TrojanServer) UpdateTraffic(upload, download int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upload += upload
	s.download += download
}

// UpdateLastActive updates last active time
func (s *TrojanServer) UpdateLastActive(time time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActiveAt = time
}

// GetLastActive returns last active time
func (s *TrojanServer) GetLastActive() time.Time {
	return s.lastActiveAt
}

// GetProtocol returns the protocol type
func (s *TrojanServer) GetProtocol() common.ProtocolType {
	return common.ProtocolTrojan
}

// accept 接受连接
func (s *TrojanServer) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
					continue
				}
				s.logger.Error("Failed to accept connection", logger.Fields{
					"error": err.Error(),
				})
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
		s.logger.Error("Failed to read header", logger.Fields{
			"error": err.Error(),
		})
		return
	}

	// Verify password hash
	if !s.verifyPassword(header.PasswordHash) {
		s.logger.Error("Invalid password hash", nil)
		return
	}

	// Connect to target
	target, err := net.Dial("tcp", header.Address)
	if err != nil {
		s.logger.Error("Failed to connect to target", logger.Fields{
			"error":   err.Error(),
			"address": header.Address,
		})
		return
	}
	defer target.Close()

	// Update last active time
	s.UpdateLastActive(time.Now())

	// Start proxying
	upload, download := TrojanCopyIO(conn, target)

	// Update traffic stats
	s.UpdateTraffic(upload, download)
}

// HandleConnection implements the common.ProxyServerInterface
func (s *TrojanServer) HandleConnection(conn io.ReadWriteCloser) error {
	// Convert to net.Conn if possible
	netConn, ok := conn.(net.Conn)
	if !ok {
		return fmt.Errorf("connection is not a net.Conn")
	}

	go s.handleConnection(netConn)
	return nil
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

// TrojanCopyIO copies data between two io.ReadWriteCloser and returns traffic stats
func TrojanCopyIO(dst io.Writer, src io.Reader) (int64, int64) {
	var upload, download int64

	// Create channels for copy results
	ch := make(chan int64, 2)

	// Copy from src to dst (upload)
	go func() {
		n, err := io.Copy(dst, src)
		if err != nil {
			// Just log or ignore the error
		}
		ch <- n
	}()

	// Copy from dst to src (download) - for our case, dst parameter is the client connection
	n, err := io.Copy(src.(io.Writer), dst.(io.Reader))
	if err != nil {
		// Just log or ignore the error
	}
	download = n

	// Get upload bytes from goroutine
	upload = <-ch

	return upload, download
}
