package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"v/common"
	"v/logger"
)

// SocksServer SOCKS 服务器
type SocksServer struct {
	logger   *logger.Logger
	proxy    *common.Proxy
	config   *common.SocksConfig
	listener net.Listener
	mu       sync.Mutex
}

// NewSocksServer 创建 SOCKS 服务器
func NewSocksServer(logger *logger.Logger, proxy *common.Proxy) (*SocksServer, error) {
	// 解析配置
	var config common.SocksConfig
	if err := json.Unmarshal([]byte(proxy.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &SocksServer{
		logger: logger,
		proxy:  proxy,
		config: &config,
	}, nil
}

// Start 启动服务器
func (s *SocksServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.proxy.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.listener = listener
	go s.accept()
	return nil
}

// Stop 停止服务器
func (s *SocksServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
		}
		s.listener = nil
	}
	return nil
}

// GetPort 获取端口
func (s *SocksServer) GetPort() int {
	return s.proxy.Port
}

// GetProtocol 获取协议类型
func (s *SocksServer) GetProtocol() common.ProtocolType {
	return common.ProtocolSocks
}

// accept 接受连接
func (s *SocksServer) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("failed to accept connection", "error", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection 处理连接
func (s *SocksServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取版本
	version := make([]byte, 1)
	if _, err := conn.Read(version); err != nil {
		s.logger.Error("failed to read version", "error", err)
		return
	}

	// 处理认证
	if version[0] == 0x05 {
		if err := s.handleSocks5(conn); err != nil {
			s.logger.Error("failed to handle socks5", "error", err)
			return
		}
	} else if version[0] == 0x04 {
		if err := s.handleSocks4(conn); err != nil {
			s.logger.Error("failed to handle socks4", "error", err)
			return
		}
	} else {
		s.logger.Error("unsupported socks version", "version", version[0])
	}
}

// handleSocks5 处理 SOCKS5 连接
func (s *SocksServer) handleSocks5(conn net.Conn) error {
	// 读取认证方法数量
	methodCount := make([]byte, 1)
	if _, err := conn.Read(methodCount); err != nil {
		return err
	}

	// 读取认证方法
	methods := make([]byte, methodCount[0])
	if _, err := conn.Read(methods); err != nil {
		return err
	}

	// 选择认证方法
	var method byte
	if s.config.Auth == "none" {
		method = 0x00
	} else if s.config.Auth == "password" {
		method = 0x02
	} else {
		method = 0xFF
	}

	// 发送认证响应
	if _, err := conn.Write([]byte{0x05, method}); err != nil {
		return err
	}

	// 处理密码认证
	if method == 0x02 {
		if err := s.handlePasswordAuth(conn); err != nil {
			return err
		}
	}

	// 读取请求
	request := make([]byte, 4)
	if _, err := conn.Read(request); err != nil {
		return err
	}

	// 处理请求
	return s.handleSocks5Request(conn, request)
}

// handleSocks4 处理 SOCKS4 连接
func (s *SocksServer) handleSocks4(conn net.Conn) error {
	// 读取请求头
	header := make([]byte, 9)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("failed to read request header: %v", err)
	}

	// 检查版本
	if header[0] != 0x04 {
		return fmt.Errorf("unsupported SOCKS version: %d", header[0])
	}

	// 获取命令
	cmd := header[1]
	if cmd != 0x01 { // CONNECT
		return fmt.Errorf("unsupported command: %d", cmd)
	}

	// 获取端口
	port := uint16(header[2])<<8 | uint16(header[3])

	// 获取目标地址
	addr := make([]byte, 4)
	copy(addr, header[4:8])

	// 读取用户ID
	userID := make([]byte, 0)
	for {
		b := make([]byte, 1)
		if _, err := conn.Read(b); err != nil {
			return fmt.Errorf("failed to read user ID: %v", err)
		}
		if b[0] == 0 {
			break
		}
		userID = append(userID, b[0])
	}

	// 连接目标服务器
	target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", net.IP(addr).String(), port))
	if err != nil {
		// 发送失败响应
		response := []byte{0x00, 0x5B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		if _, err := conn.Write(response); err != nil {
			return fmt.Errorf("failed to write response: %v", err)
		}
		return fmt.Errorf("failed to connect to target: %v", err)
	}
	defer target.Close()

	// 发送成功响应
	response := []byte{0x00, 0x5A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if _, err := conn.Write(response); err != nil {
		return fmt.Errorf("failed to write response: %v", err)
	}

	// 开始转发数据
	go CopyIO(conn, target)
	CopyIO(target, conn)

	return nil
}

// handlePasswordAuth 处理密码认证
func (s *SocksServer) handlePasswordAuth(conn net.Conn) error {
	// 读取版本
	version := make([]byte, 1)
	if _, err := conn.Read(version); err != nil {
		return err
	}

	// 读取用户名长度
	userLen := make([]byte, 1)
	if _, err := conn.Read(userLen); err != nil {
		return err
	}

	// 读取用户名
	username := make([]byte, userLen[0])
	if _, err := conn.Read(username); err != nil {
		return err
	}

	// 读取密码长度
	passLen := make([]byte, 1)
	if _, err := conn.Read(passLen); err != nil {
		return err
	}

	// 读取密码
	password := make([]byte, passLen[0])
	if _, err := conn.Read(password); err != nil {
		return err
	}

	// 验证用户名和密码
	if string(username) != s.config.Username || string(password) != s.config.Password {
		if _, err := conn.Write([]byte{0x01, 0x01}); err != nil {
			return err
		}
		return fmt.Errorf("invalid username or password")
	}

	// 发送认证成功响应
	if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
		return err
	}

	return nil
}

// handleSocks5Request 处理 SOCKS5 请求
func (s *SocksServer) handleSocks5Request(conn net.Conn, request []byte) error {
	// 读取目标地址
	var addr string
	var port uint16

	switch request[3] {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := conn.Read(ip); err != nil {
			return err
		}
		addr = net.IP(ip).String()

	case 0x03: // Domain
		domainLen := make([]byte, 1)
		if _, err := conn.Read(domainLen); err != nil {
			return err
		}
		domain := make([]byte, domainLen[0])
		if _, err := conn.Read(domain); err != nil {
			return err
		}
		addr = string(domain)

	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := conn.Read(ip); err != nil {
			return err
		}
		addr = net.IP(ip).String()

	default:
		return fmt.Errorf("unsupported address type: %d", request[3])
	}

	// 读取端口
	portBytes := make([]byte, 2)
	if _, err := conn.Read(portBytes); err != nil {
		return err
	}
	port = uint16(portBytes[0])<<8 | uint16(portBytes[1])

	// 连接到目标服务器
	target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		// 发送连接失败响应
		response := []byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0}
		if _, err := conn.Write(response); err != nil {
			return err
		}
		return err
	}
	defer target.Close()

	// 发送连接成功响应
	localAddr := target.LocalAddr().(*net.TCPAddr)
	response := []byte{0x05, 0x00, 0x00, 0x01}
	response = append(response, localAddr.IP.To4()...)
	response = append(response, byte(localAddr.Port>>8), byte(localAddr.Port))
	if _, err := conn.Write(response); err != nil {
		return err
	}

	// 开始转发数据
	go CopyIO(conn, target)
	CopyIO(target, conn)

	return nil
}

// HandleConnection 实现 Server 接口
func (s *SocksServer) HandleConnection(conn io.ReadWriteCloser) error {
	// 将 io.ReadWriteCloser 转换为 net.Conn
	netConn, ok := conn.(net.Conn)
	if !ok {
		return fmt.Errorf("connection is not a net.Conn")
	}

	// 处理连接
	s.handleConnection(netConn)
	return nil
}
