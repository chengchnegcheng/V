package vless

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"time"

	"v/common"
	"v/logger"
)

// Config represents VLESS configuration
type Config struct {
	ID       string `json:"id"`
	Flow     string `json:"flow,omitempty"`
	Security string `json:"security,omitempty"`
}

// Server represents a VLESS server
type Server struct {
	log    *logger.Logger
	config *Config
	proxy  *common.ProxyInstance
	aead   cipher.AEAD
	block  cipher.Block
}

// New creates a new VLESS server
func New(log *logger.Logger, proxy *common.ProxyInstance) (*Server, error) {
	var vlessConfig map[string]interface{}

	// Handle different types of settings
	settingsData := proxy.Settings["vless"]

	switch v := settingsData.(type) {
	case map[string]interface{}:
		vlessConfig = v
	default:
		return nil, fmt.Errorf("invalid vless config type: %T", settingsData)
	}

	id, ok := vlessConfig["id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing vless id")
	}

	flow, ok := vlessConfig["flow"].(string)
	if !ok {
		flow = ""
	}

	security, ok := vlessConfig["security"].(string)
	if !ok {
		security = "none"
	}

	// Create config struct
	config := &Config{
		ID:       id,
		Flow:     flow,
		Security: security,
	}

	// Use AES-128 for encryption
	key := sha256.Sum256([]byte(id))
	block, err := aes.NewCipher(key[:16])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create AEAD instance
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD: %v", err)
	}

	s := &Server{
		log:    log,
		config: config,
		proxy:  proxy,
		block:  block,
		aead:   aead,
	}

	return s, nil
}

// Start starts the VLESS server
func (s *Server) Start() error {
	s.log.Info("Starting VLESS server", logger.Fields{
		"proxy_id": s.proxy.ID,
		"port":     s.proxy.Port,
	})

	// TODO: Implement server start logic

	return nil
}

// Stop stops the VLESS server
func (s *Server) Stop() error {
	s.log.Info("Stopping VLESS server", logger.Fields{
		"proxy_id": s.proxy.ID,
		"port":     s.proxy.Port,
	})

	// TODO: Implement server stop logic

	return nil
}

// HandleConnection handles a new connection
func (s *Server) HandleConnection(conn io.ReadWriteCloser) error {
	// Read request header
	header, err := s.readHeader(conn)
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}

	// Validate request
	if err := s.validateRequest(header); err != nil {
		return fmt.Errorf("invalid request: %v", err)
	}

	// Handle request
	switch header.Command {
	case 1: // TCP
		return s.handleTCP(conn, header)
	case 2: // UDP
		return s.handleUDP(conn, header)
	default:
		return fmt.Errorf("unsupported command: %d", header.Command)
	}
}

// Header represents VLESS request header
type Header struct {
	Version     byte
	Command     byte
	Option      byte
	Port        uint16
	AddressType byte
	Address     string
	Timestamp   int64
}

// readHeader reads and decodes the VLESS request header
func (s *Server) readHeader(conn io.Reader) (*Header, error) {
	// Read encrypted header
	buf := make([]byte, 16)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	// Decrypt header
	nonce := make([]byte, 12)
	copy(nonce, buf[:12])
	plaintext, err := s.aead.Open(nil, nonce, buf[12:], nil)
	if err != nil {
		return nil, err
	}

	// Parse header
	header := &Header{
		Version:   plaintext[0],
		Command:   plaintext[1],
		Option:    plaintext[2],
		Port:      uint16(plaintext[3])<<8 | uint16(plaintext[4]),
		Timestamp: time.Now().Unix(),
	}

	// Read address
	addrLen := plaintext[5]
	addrBuf := make([]byte, addrLen)
	if _, err := io.ReadFull(conn, addrBuf); err != nil {
		return nil, err
	}

	header.AddressType = addrBuf[0]
	header.Address = string(addrBuf[1:])

	return header, nil
}

// validateRequest validates the VLESS request
func (s *Server) validateRequest(header *Header) error {
	// Check version
	if header.Version != 0 {
		return fmt.Errorf("unsupported version: %d", header.Version)
	}

	// Check timestamp
	now := time.Now().Unix()
	if now-header.Timestamp > 120 || header.Timestamp-now > 120 {
		return fmt.Errorf("invalid timestamp")
	}

	// Check address type
	switch header.AddressType {
	case 1: // IPv4
	case 2: // Domain
	case 3: // IPv6
	default:
		return fmt.Errorf("unsupported address type: %d", header.AddressType)
	}

	return nil
}

// handleTCP handles TCP connection
func (s *Server) handleTCP(conn io.ReadWriteCloser, header *Header) error {
	// Connect to target address
	target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", header.Address, header.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to target: %v", err)
	}
	defer target.Close()

	// Start proxying data
	go func() {
		io.Copy(target, conn)
	}()
	io.Copy(conn, target)

	return nil
}

// handleUDP handles UDP connection
func (s *Server) handleUDP(conn io.ReadWriteCloser, header *Header) error {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", header.Address, header.Port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	// Create UDP connection
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer udpConn.Close()

	// Create buffer for UDP packets
	buf := make([]byte, 65536)

	// Start proxying UDP data
	// This is a simplified implementation
	go func() {
		for {
			n, err := udpConn.Read(buf)
			if err != nil {
				break
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				break
			}
		}
	}()

	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		_, err = udpConn.Write(buf[:n])
		if err != nil {
			break
		}
	}

	return nil
}
