package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

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
	Port       int
	Upload     int64
	Download   int64
	LastActive time.Time
	Running    bool
	Listener   net.Listener
	config     *DokodemoConfig
	log        *logger.Logger
	mu         sync.Mutex
}

// NewDokodemoServer 创建 Dokodemo 服务器
func NewDokodemoServer(logger *logger.Logger, proxy *common.ProxyConfig) (*DokodemoServer, error) {
	// 解析配置
	var config DokodemoConfig
	if err := json.Unmarshal([]byte(proxy.Settings), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &DokodemoServer{
		Port:   proxy.Port,
		config: &config,
		log:    logger,
	}, nil
}

// Start 启动服务器
func (s *DokodemoServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.Listener = listener
	go s.accept()
	return nil
}

// Stop 停止服务器
func (s *DokodemoServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Listener != nil {
		if err := s.Listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
		}
		s.Listener = nil
	}
	return nil
}

// GetPort 获取端口
func (s *DokodemoServer) GetPort() int {
	return s.Port
}

// GetProtocol 获取协议类型
func (s *DokodemoServer) GetProtocol() common.ProtocolType {
	return common.ProtocolDokodemo
}

// accept 接受连接
func (s *DokodemoServer) accept() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			s.log.Error("failed to accept connection", "error", err)
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
		s.log.Error("failed to connect to target", "error", err)
		return
	}
	defer target.Close()

	// 开始转发数据
	go DokoCopyIO(conn, target)
	DokoCopyIO(target, conn)
}

// DokoCopyIO copies data between two connections
func DokoCopyIO(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()
	io.Copy(dst, src)
}
