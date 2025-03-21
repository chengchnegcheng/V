package proxy

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"v/common"
	"v/logger"
)

// VLESSConfig VLESS 配置
type VLESSConfig struct {
	ID       string `json:"id"`
	Flow     string `json:"flow"`
	Security string `json:"security"`
}

// VLESServer VLESS 服务器
type VLESServer struct {
	logger   *logger.Logger
	proxy    *common.Proxy
	config   *VLESSConfig
	listener net.Listener
	mu       sync.Mutex
}

// NewVLESServer 创建 VLESS 服务器
func NewVLESServer(logger *logger.Logger, proxy *common.Proxy) (*VLESServer, error) {
	// 解析配置
	var config VLESSConfig
	if err := json.Unmarshal([]byte(proxy.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &VLESServer{
		logger: logger,
		proxy:  proxy,
		config: &config,
	}, nil
}

// Start 启动服务器
func (s *VLESServer) Start() error {
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
func (s *VLESServer) Stop() error {
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
func (s *VLESServer) GetPort() int {
	return s.proxy.Port
}

// GetProtocol 获取协议类型
func (s *VLESServer) GetProtocol() common.ProtocolType {
	return common.ProtocolVLESS
}

// accept 接受连接
func (s *VLESServer) accept() {
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
func (s *VLESServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// TODO: 实现 VLESS 协议处理
	s.logger.Info("new connection", "remote", conn.RemoteAddr())
}
