package proxy

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"v/common"
	"v/logger"
)

// DokodemoConfig Dokodemo 配置
type DokodemoConfig struct {
	TargetAddr string `json:"target_addr"`
	TargetPort int    `json:"target_port"`
	Network    string `json:"network"`
	Timeout    int    `json:"timeout"`
}

// DokodemoServer Dokodemo 服务器
type DokodemoServer struct {
	logger   *logger.Logger
	proxy    *common.Proxy
	config   *DokodemoConfig
	listener net.Listener
	mu       sync.Mutex
}

// NewDokodemoServer 创建 Dokodemo 服务器
func NewDokodemoServer(logger *logger.Logger, proxy *common.Proxy) (*DokodemoServer, error) {
	// 解析配置
	var config DokodemoConfig
	if err := json.Unmarshal([]byte(proxy.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &DokodemoServer{
		logger: logger,
		proxy:  proxy,
		config: &config,
	}, nil
}

// Start 启动服务器
func (s *DokodemoServer) Start() error {
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
func (s *DokodemoServer) Stop() error {
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
func (s *DokodemoServer) GetPort() int {
	return s.proxy.Port
}

// GetProtocol 获取协议类型
func (s *DokodemoServer) GetProtocol() common.ProtocolType {
	return common.ProtocolDokodemo
}

// accept 接受连接
func (s *DokodemoServer) accept() {
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
func (s *DokodemoServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 连接到目标服务器
	target, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.config.TargetAddr, s.config.TargetPort))
	if err != nil {
		s.logger.Error("failed to connect to target", "error", err)
		return
	}
	defer target.Close()

	// 开始转发数据
	go CopyIO(conn, target)
	CopyIO(target, conn)
}

// CopyIO 复制数据
func CopyIO(dst net.Conn, src net.Conn) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if err != nil {
			return
		}

		if n > 0 {
			if _, err := dst.Write(buf[:n]); err != nil {
				return
			}
		}
	}
}
