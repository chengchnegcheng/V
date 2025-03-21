package trojan

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net"

	"v/logger"
	"v/proxy"
)

// Config represents Trojan configuration
type Config struct {
	Password string `json:"password"`
	SSL      struct {
		Cert string `json:"cert"`
		Key  string `json:"key"`
	} `json:"ssl"`
}

// Server represents a Trojan server
type Server struct {
	log      *logger.Logger
	config   *Config
	proxy    *proxy.Proxy
	password []byte
}

// New creates a new Trojan server
func New(log *logger.Logger, proxy *proxy.Proxy) (*Server, error) {
	config, ok := proxy.Config.Settings["trojan"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid trojan config")
	}

	password, ok := config["password"].(string)
	if !ok {
		return nil, fmt.Errorf("missing trojan password")
	}

	ssl, ok := config["ssl"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing trojan ssl config")
	}

	cert, ok := ssl["cert"].(string)
	if !ok {
		return nil, fmt.Errorf("missing trojan ssl cert")
	}

	key, ok := ssl["key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing trojan ssl key")
	}

	trojanConfig := &Config{
		Password: password,
		SSL: struct {
			Cert string `json:"cert"`
			Key  string `json:"key"`
		}{
			Cert: cert,
			Key:  key,
		},
	}

	// Hash password
	hash := sha256.New()
	hash.Write([]byte(password))
	hashedPassword := hash.Sum(nil)

	return &Server{
		log:      log,
		config:   trojanConfig,
		proxy:    proxy,
		password: hashedPassword,
	}, nil
}

// Start starts the Trojan server
func (s *Server) Start() error {
	s.log.Info("Starting Trojan server", logger.Fields{
		"proxy_id": s.proxy.ID,
		"port":     s.proxy.Port,
	})

	// TODO: Implement server start logic

	return nil
}

// Stop stops the Trojan server
func (s *Server) Stop() error {
	s.log.Info("Stopping Trojan server", logger.Fields{
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

// Header represents Trojan request header
type Header struct {
	Hash        []byte
	Command     byte
	Port        uint16
	AddressType byte
	Address     string
}

// readHeader reads and decodes the Trojan request header
func (s *Server) readHeader(conn io.Reader) (*Header, error) {
	// Read hash
	hash := make([]byte, 56)
	if _, err := io.ReadFull(conn, hash); err != nil {
		return nil, err
	}

	// Read command
	cmd := make([]byte, 1)
	if _, err := io.ReadFull(conn, cmd); err != nil {
		return nil, err
	}

	// Read port
	port := make([]byte, 2)
	if _, err := io.ReadFull(conn, port); err != nil {
		return nil, err
	}

	// Read address type
	addrType := make([]byte, 1)
	if _, err := io.ReadFull(conn, addrType); err != nil {
		return nil, err
	}

	// Read address
	var addr string
	switch addrType[0] {
	case 1: // IPv4
		addrBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn, addrBuf); err != nil {
			return nil, err
		}
		addr = net.IP(addrBuf).String()
	case 2: // Domain
		domainLen := make([]byte, 1)
		if _, err := io.ReadFull(conn, domainLen); err != nil {
			return nil, err
		}
		domainBuf := make([]byte, domainLen[0])
		if _, err := io.ReadFull(conn, domainBuf); err != nil {
			return nil, err
		}
		addr = string(domainBuf)
	case 3: // IPv6
		addrBuf := make([]byte, 16)
		if _, err := io.ReadFull(conn, addrBuf); err != nil {
			return nil, err
		}
		addr = net.IP(addrBuf).String()
	default:
		return nil, fmt.Errorf("unsupported address type: %d", addrType[0])
	}

	return &Header{
		Hash:        hash,
		Command:     cmd[0],
		Port:        uint16(port[0])<<8 | uint16(port[1]),
		AddressType: addrType[0],
		Address:     addr,
	}, nil
}

// validateRequest validates the Trojan request
func (s *Server) validateRequest(header *Header) error {
	// Check hash
	if !bytes.Equal(header.Hash, s.password) {
		return fmt.Errorf("invalid password")
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
	// Connect to target
	target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", header.Address, header.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to target: %v", err)
	}
	defer target.Close()

	// Start proxying
	go io.Copy(target, conn)
	io.Copy(conn, target)

	return nil
}

// handleUDP handles UDP connection
func (s *Server) handleUDP(conn io.ReadWriteCloser, header *Header) error {
	// Create UDP connection
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", header.Address, header.Port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer udpConn.Close()

	// Create buffer for UDP packets
	buf := make([]byte, 65536)
	for {
		// Read from client
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read from client: %v", err)
		}

		// Send to target
		_, err = udpConn.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("failed to write to target: %v", err)
		}

		// Read from target
		n, _, err = udpConn.ReadFromUDP(buf)
		if err != nil {
			return fmt.Errorf("failed to read from target: %v", err)
		}

		// Send back to client
		_, err = conn.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("failed to write to client: %v", err)
		}
	}
}
