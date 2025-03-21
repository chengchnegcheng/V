package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"v/common"
	"v/logger"
)

// HTTPServer HTTP 服务器
type HTTPServer struct {
	logger *logger.Logger
	proxy  *common.Proxy
	config *common.HTTPConfig
	server *http.Server
	mu     sync.Mutex
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(logger *logger.Logger, proxy *common.Proxy) (*HTTPServer, error) {
	// 解析配置
	var config common.HTTPConfig
	if err := json.Unmarshal([]byte(proxy.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &HTTPServer{
		logger: logger,
		proxy:  proxy,
		config: &config,
	}, nil
}

// Start 启动服务器
func (s *HTTPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建 HTTP 服务器
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.proxy.Port),
		Handler: s,
	}

	// 启动服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("failed to start server", "error", err)
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *HTTPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		if err := s.server.Close(); err != nil {
			return fmt.Errorf("failed to close server: %v", err)
		}
		s.server = nil
	}
	return nil
}

// GetPort 获取端口
func (s *HTTPServer) GetPort() int {
	return s.proxy.Port
}

// GetProtocol 获取协议类型
func (s *HTTPServer) GetProtocol() common.ProtocolType {
	return common.ProtocolHTTP
}

// ServeHTTP 实现 http.Handler 接口
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 处理认证
	if s.config.Auth == "basic" {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if username != s.config.Username || password != s.config.Password {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// 处理 CONNECT 方法
	if r.Method == http.MethodConnect {
		s.handleConnect(w, r)
		return
	}

	// 处理普通 HTTP 请求
	s.handleHTTP(w, r)
}

// handleConnect 处理 CONNECT 请求
func (s *HTTPServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	// 获取目标地址
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		http.Error(w, "Invalid host", http.StatusBadRequest)
		return
	}

	// 连接到目标服务器
	target, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		http.Error(w, "Failed to connect to target", http.StatusBadGateway)
		return
	}
	defer target.Close()

	// 发送连接成功响应
	w.WriteHeader(http.StatusOK)

	// 获取底层连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// 开始转发数据
	go CopyIO(conn, target)
	CopyIO(target, conn)
}

// handleHTTP 处理普通 HTTP 请求
func (s *HTTPServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建请求
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	io.Copy(w, resp.Body)
}

// HandleConnection 实现 Server 接口
func (s *HTTPServer) HandleConnection(conn io.ReadWriteCloser) error {
	// HTTP 服务器使用 ServeHTTP 方法处理连接
	return fmt.Errorf("HTTP server does not support direct connection handling")
}
