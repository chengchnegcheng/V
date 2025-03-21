package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
	"v/database"
	"v/logger"
	"v/model"

	"v/common"
)

var (
	ErrInvalidProtocol = errors.New("invalid protocol")
	ErrPortInUse       = errors.New("port is already in use")
	ErrInvalidSettings = errors.New("invalid proxy settings")
	ErrProxyNotFound   = errors.New("proxy not found")
	ErrProxyDisabled   = errors.New("proxy is disabled")
)

// DefaultService is the global proxy service instance
var DefaultService = NewProxyService()

// DefaultLimiter is the global traffic limiter instance
var DefaultLimiter = NewTrafficLimiter()

// ProxyService manages all proxy servers
type ProxyService struct {
	sync.RWMutex
	servers map[int64]*common.ProxyServer
}

// ProxyServerInterface defines the interface for proxy servers
type ProxyServerInterface interface {
	Start() error
	Stop() error
	HandleConnection(conn io.ReadWriteCloser) error
}

// NewProxyServer creates a new proxy server instance
func NewProxyServer(logger *logger.Logger, proxy *model.Proxy) (*common.ProxyServer, error) {
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(proxy.Settings), &settings); err != nil {
		return nil, fmt.Errorf("invalid settings format: %v", err)
	}

	server := &common.ProxyServer{
		ID:           proxy.ID,
		UserID:       proxy.UserID,
		Protocol:     string(proxy.Protocol),
		Port:         proxy.Port,
		Settings:     settings,
		Enabled:      proxy.Enabled,
		Upload:       proxy.Upload,
		Download:     proxy.Download,
		LastActiveAt: proxy.LastActiveAt,
		CreatedAt:    proxy.CreatedAt,
		UpdatedAt:    proxy.UpdatedAt,
	}

	// 创建服务器实例
	var err error
	switch proxy.Protocol {
	case model.ProtocolVMess:
		config := &common.VMessConfig{
			ID:       settings["id"].(string),
			Security: settings["security"].(string),
		}
		server.Server, err = NewVMessServer(logger, config, server)
	case model.ProtocolVLESS:
		config := &common.VLESSConfig{
			ID:       settings["id"].(string),
			Flow:     settings["flow"].(string),
			Security: settings["security"].(string),
		}
		server.Server, err = NewVLESSServer(logger, config, server)
	case model.ProtocolTrojan:
		config := &common.TrojanConfig{
			Password: settings["password"].(string),
			Security: settings["security"].(string),
		}
		server.Server, err = NewTrojanServer(logger, config, server)
	case model.ProtocolShadowsocks:
		config := &common.ShadowsocksConfig{
			Method:   settings["method"].(string),
			Password: settings["password"].(string),
			Security: settings["security"].(string),
		}
		server.Server, err = NewShadowsocksServer(logger, config, server)
	default:
		return nil, ErrInvalidProtocol
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create server: %v", err)
	}

	return server, nil
}

// NewProxyService creates a new proxy service
func NewProxyService() *ProxyService {
	return &ProxyService{
		servers: make(map[int64]*common.ProxyServer),
	}
}

// InitService initializes the proxy service
func InitService() error {
	// Load all proxy configurations from database
	proxies, err := database.DBInstance.GetAllProxies()
	if err != nil {
		return fmt.Errorf("failed to load proxy configurations: %v", err)
	}

	for _, proxy := range proxies {
		if !proxy.Enabled {
			continue
		}

		// Create and start proxy server
		server, err := NewProxyServer(logger.NewLogger(), proxy)
		if err != nil {
			log.Printf("Failed to create proxy server: %v", err)
			continue
		}

		if err := server.Server.Start(); err != nil {
			log.Printf("Failed to start proxy server: %v", err)
			continue
		}

		DefaultService.servers[proxy.ID] = server
	}

	// Start traffic statistics collector
	go DefaultService.collectTrafficStats()

	return nil
}

// collectTrafficStats collects traffic statistics periodically
func (s *ProxyService) collectTrafficStats() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateTrafficStats()
		}
	}
}

// updateTrafficStats updates traffic statistics for all servers
func (s *ProxyService) updateTrafficStats() {
	s.RLock()
	defer s.RUnlock()

	for id, server := range s.servers {
		proxy := &model.Proxy{
			ID:       id,
			Upload:   server.Upload,
			Download: server.Download,
		}
		if err := database.DBInstance.UpdateProxyStats(proxy); err != nil {
			log.Printf("Failed to update traffic stats: %v", err)
		}
	}
}

// Create creates a new proxy
func (s *ProxyService) Create(proxy *model.Proxy) error {
	s.Lock()
	defer s.Unlock()

	// Check if port is available
	for _, server := range s.servers {
		if server.Port == proxy.Port {
			return ErrPortInUse
		}
	}

	// Create proxy in database
	if err := database.DBInstance.CreateProxy(proxy); err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	// Create and start proxy server
	server, err := NewProxyServer(logger.NewLogger(), proxy)
	if err != nil {
		return fmt.Errorf("failed to create proxy server: %v", err)
	}

	if err := server.Server.Start(); err != nil {
		return fmt.Errorf("failed to start proxy server: %v", err)
	}

	s.servers[proxy.ID] = server
	return nil
}

// Get returns a proxy by ID
func (s *ProxyService) Get(id int64) (*common.ProxyServer, error) {
	s.RLock()
	defer s.RUnlock()

	server, exists := s.servers[id]
	if !exists {
		return nil, ErrProxyNotFound
	}

	return server, nil
}

// GetByUser returns all proxies for a user
func (s *ProxyService) GetByUser(userID int64) ([]*common.ProxyServer, error) {
	s.RLock()
	defer s.RUnlock()

	var userServers []*common.ProxyServer
	for _, server := range s.servers {
		if server.UserID == userID {
			userServers = append(userServers, server)
		}
	}

	return userServers, nil
}

// Update updates a proxy
func (s *ProxyService) Update(proxy *model.Proxy) error {
	s.Lock()
	defer s.Unlock()

	server, exists := s.servers[proxy.ID]
	if !exists {
		return ErrProxyNotFound
	}

	// Stop the server if it's running
	if err := server.Server.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %v", err)
	}

	// Update proxy in database
	if err := database.DBInstance.UpdateProxy(proxy); err != nil {
		return fmt.Errorf("failed to update proxy: %v", err)
	}

	// Create and start new server
	newServer, err := NewProxyServer(logger.NewLogger(), proxy)
	if err != nil {
		return fmt.Errorf("failed to create new server: %v", err)
	}

	if err := newServer.Server.Start(); err != nil {
		return fmt.Errorf("failed to start new server: %v", err)
	}

	s.servers[proxy.ID] = newServer
	return nil
}

// Delete deletes a proxy
func (s *ProxyService) Delete(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return ErrProxyNotFound
	}

	// Stop the server
	if err := server.Server.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %v", err)
	}

	// Delete proxy from database
	if err := s.db.DeleteProxy(id); err != nil {
		return fmt.Errorf("failed to delete proxy: %v", err)
	}

	delete(s.servers, id)
	return nil
}

// List returns all proxies
func (s *ProxyService) List() []*common.ProxyServer {
	s.RLock()
	defer s.RUnlock()

	servers := make([]*common.ProxyServer, 0, len(s.servers))
	for _, server := range s.servers {
		servers = append(servers, server)
	}

	return servers
}

// UpdateTraffic updates traffic statistics for a proxy
func (s *ProxyService) UpdateTraffic(id int64, upload, download int64) error {
	s.Lock()
	defer s.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return ErrProxyNotFound
	}

	server.Upload += upload
	server.Download += download

	proxy := &model.Proxy{
		ID:       id,
		Upload:   server.Upload,
		Download: server.Download,
	}

	if err := database.DBInstance.UpdateProxyStats(proxy); err != nil {
		return fmt.Errorf("failed to update traffic stats: %v", err)
	}

	return nil
}

// Enable enables a proxy
func (s *ProxyService) Enable(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return ErrProxyNotFound
	}

	proxy := &model.Proxy{
		ID:      id,
		Enabled: true,
	}

	if err := database.DBInstance.UpdateProxy(proxy); err != nil {
		return fmt.Errorf("failed to enable proxy: %v", err)
	}

	if err := server.Server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	return nil
}

// Disable disables a proxy
func (s *ProxyService) Disable(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return ErrProxyNotFound
	}

	proxy := &model.Proxy{
		ID:      id,
		Enabled: false,
	}

	if err := database.DBInstance.UpdateProxy(proxy); err != nil {
		return fmt.Errorf("failed to disable proxy: %v", err)
	}

	if err := server.Server.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %v", err)
	}

	return nil
}

// UpdateLastActive updates the last active time for a proxy
func (s *ProxyService) UpdateLastActive(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, exists := s.servers[id]
	if !exists {
		return ErrProxyNotFound
	}

	proxy := &model.Proxy{
		ID:           id,
		LastActiveAt: time.Now(),
	}

	if err := database.DBInstance.UpdateProxy(proxy); err != nil {
		return fmt.Errorf("failed to update last active time: %v", err)
	}

	server.LastActiveAt = time.Now()
	return nil
}

// ProxyServerManager 代理服务器管理器
type ProxyServerManager struct {
	servers map[int64]*common.ProxyServer
	mu      sync.RWMutex
}

// NewProxyServerManager 创建代理服务器管理器
func NewProxyServerManager() *ProxyServerManager {
	return &ProxyServerManager{
		servers: make(map[int64]*common.ProxyServer),
	}
}

// AddServer 添加代理服务器
func (m *ProxyServerManager) AddServer(proxy *common.ProxyServer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[proxy.ID]; exists {
		return fmt.Errorf("server with ID %d already exists", proxy.ID)
	}

	m.servers[proxy.ID] = proxy
	return nil
}

// RemoveServer 移除代理服务器
func (m *ProxyServerManager) RemoveServer(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[id]; !exists {
		return fmt.Errorf("server with ID %d does not exist", id)
	}

	delete(m.servers, id)
	return nil
}

// GetServer 获取代理服务器
func (m *ProxyServerManager) GetServer(id int64) (*common.ProxyServer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[id]
	if !exists {
		return nil, fmt.Errorf("server with ID %d does not exist", id)
	}

	return server, nil
}

// GetAllServers 获取所有代理服务器
func (m *ProxyServerManager) GetAllServers() []*common.ProxyServer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]*common.ProxyServer, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}

	return servers
}

// StartServer 启动代理服务器
func (m *ProxyServerManager) StartServer(id int64) error {
	server, err := m.GetServer(id)
	if err != nil {
		return err
	}

	return server.Server.Start()
}

// StopServer 停止代理服务器
func (m *ProxyServerManager) StopServer(id int64) error {
	server, err := m.GetServer(id)
	if err != nil {
		return err
	}

	return server.Server.Stop()
}

// GetServerStats 获取代理服务器统计信息
func (m *ProxyServerManager) GetServerStats(id int64) (*common.ProxyStats, error) {
	server, err := m.GetServer(id)
	if err != nil {
		return nil, err
	}

	return &common.ProxyStats{
		Upload:    server.Upload,
		Download:  server.Download,
		Timestamp: server.LastActiveAt.Unix(),
	}, nil
}

// UpdateServerStats 更新代理服务器统计信息
func (m *ProxyServerManager) UpdateServerStats(id int64, upload, download int64) error {
	server, err := m.GetServer(id)
	if err != nil {
		return err
	}

	server.Upload += upload
	server.Download += download
	return nil
}
