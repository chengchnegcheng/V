package proxy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"net"
	"time"

	"v/common"
	"v/logger"
)

// ShadowsocksServer represents a Shadowsocks server
type ShadowsocksServer struct {
	*BaseServer
	config *common.ShadowsocksConfig
	cipher cipher.Block
}

// NewShadowsocksServer creates a new Shadowsocks server
func NewShadowsocksServer(logger *logger.Logger, config *common.ShadowsocksConfig, proxy *common.ProxyInstance) (*ShadowsocksServer, error) {
	server := &ShadowsocksServer{
		BaseServer: NewBaseServer(logger, proxy),
		config:     config,
	}

	// Initialize cipher
	key := generateKey(config.Password, config.Method)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}
	server.cipher = block

	return server, nil
}

// Start starts the Shadowsocks server
func (s *ShadowsocksServer) Start() error {
	if s.Running {
		return fmt.Errorf("server is already running")
	}

	addr := fmt.Sprintf(":%d", s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}

	s.Listener = listener
	s.Running = true
	s.Logger.Info("Shadowsocks server started on port %d", s.Port)

	// Handle connections
	go s.handleConnections()

	return nil
}

// Stop stops the Shadowsocks server
func (s *ShadowsocksServer) Stop() error {
	if !s.Running {
		return nil
	}

	if err := s.Listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %v", err)
	}

	s.Running = false
	s.Logger.Info("Shadowsocks server stopped")
	return nil
}

// handleConnections handles incoming connections
func (s *ShadowsocksServer) handleConnections() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(time.Second)
				continue
			}
			return
		}

		go s.HandleConnection(conn)
	}
}

// HandleConnection handles a single connection
func (s *ShadowsocksServer) HandleConnection(conn io.ReadWriteCloser) error {
	defer conn.Close()

	// Convert to net.Conn
	netConn, ok := conn.(net.Conn)
	if !ok {
		return fmt.Errorf("connection is not a net.Conn")
	}

	// Create stream cipher
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(netConn, iv); err != nil {
		s.Logger.Error("Failed to read IV: %v", err)
		return err
	}

	stream := cipher.NewCFBDecrypter(s.cipher, iv)
	reader := &cipher.StreamReader{S: stream, R: netConn}

	// Read address
	header, err := s.readHeader(reader)
	if err != nil {
		s.Logger.Error("Failed to read header: %v", err)
		return err
	}

	// Connect to target
	target, err := net.Dial("tcp", header.Address)
	if err != nil {
		s.Logger.Error("Failed to connect to target: %v", err)
		return err
	}
	defer target.Close()

	// Create stream cipher for response
	respIV := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(netConn, respIV); err != nil {
		s.Logger.Error("Failed to read response IV: %v", err)
		return err
	}

	respStream := cipher.NewCFBEncrypter(s.cipher, respIV)
	writer := &cipher.StreamWriter{S: respStream, W: netConn}

	// Start proxying
	errChan := make(chan error, 2)
	go func() {
		_, err := io.Copy(target, reader)
		errChan <- err
	}()
	go func() {
		_, err := io.Copy(writer, target)
		errChan <- err
	}()

	// Wait for either direction to finish
	err = <-errChan
	if err != nil {
		s.Logger.Error("Copy error: %v", err)
	}
	return err
}

// ShadowsocksHeader represents a Shadowsocks header
type ShadowsocksHeader struct {
	AddressType byte
	Address     string
	Port        uint16
}

// readHeader reads and parses the Shadowsocks header
func (s *ShadowsocksServer) readHeader(reader io.Reader) (*ShadowsocksHeader, error) {
	// Read address type
	addrType := make([]byte, 1)
	if _, err := io.ReadFull(reader, addrType); err != nil {
		return nil, fmt.Errorf("failed to read address type: %v", err)
	}

	// Read address
	var addr string
	switch addrType[0] {
	case 1: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(reader, ip); err != nil {
			return nil, fmt.Errorf("failed to read IPv4 address: %v", err)
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
	case 3: // Domain
		lenByte := make([]byte, 1)
		if _, err := io.ReadFull(reader, lenByte); err != nil {
			return nil, fmt.Errorf("failed to read domain length: %v", err)
		}
		domain := make([]byte, lenByte[0])
		if _, err := io.ReadFull(reader, domain); err != nil {
			return nil, fmt.Errorf("failed to read domain: %v", err)
		}
		addr = string(domain)
	case 4: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(reader, ip); err != nil {
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
	if _, err := io.ReadFull(reader, portBytes); err != nil {
		return nil, fmt.Errorf("failed to read port: %v", err)
	}
	port := uint16(portBytes[0])<<8 | uint16(portBytes[1])

	return &ShadowsocksHeader{
		AddressType: addrType[0],
		Address:     fmt.Sprintf("%s:%d", addr, port),
		Port:        port,
	}, nil
}

// generateKey generates a key from password and method
func generateKey(password string, method string) []byte {
	// For now, we only support AES-256-CFB
	if method != "aes-256-cfb" {
		return nil
	}

	// Generate key using EVP_BytesToKey
	md5Sum := md5.Sum([]byte(password))
	key := md5Sum[:]
	for len(key) < 32 {
		h := sha1.New()
		h.Write(key)
		h.Write([]byte(password))
		key = h.Sum(key)
	}
	return key[:32]
}
