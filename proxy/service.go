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
var DefaultService = NewService(logger.NewLogger())

// DefaultLimiter is the global traffic limiter instance
var DefaultLimiter = NewTrafficLimiter()

// Service represents a proxy service
type Service struct {
	sync.RWMutex
	logger  *logger.Logger
	servers map[int64]common.ProxyServerInterface
}

// ProxyServerInterface defines the interface for proxy servers
type ProxyServerInterface interface {
	Start() error
	Stop() error
	HandleConnection(conn io.ReadWriteCloser) error
}

// NewService creates a new proxy service
func NewService(logger *logger.Logger) *Service {
	return &Service{
		logger:  logger,
		servers: make(map[int64]common.ProxyServerInterface),
	}
}

// NewProxyServer creates a new proxy server based on the protocol type
func NewProxyServer(logger *logger.Logger, proxy *common.ProxyConfig) (common.ProxyServerInterface, error) {
	var server common.ProxyServerInterface
	var err error

	// Parse settings
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(proxy.Settings), &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %v", err)
	}

	proxyServer := &common.ProxyInstance{
		Type:     proxy.Type,
		Port:     proxy.Port,
		Settings: settings,
	}

	switch proxy.Type {
	case "vmess":
		server, err = NewVMessServer(logger, &common.VMessConfig{
			ID: settings["id"].(string),
		}, proxyServer)
	case "vless":
		server, err = NewVLESSServer(logger, &common.VLESSConfig{
			ID: settings["id"].(string),
		}, proxyServer)
	case "trojan":
		// Create Trojan config from settings
		trojanSettings, ok := settings["trojan"]
		if !ok {
			return nil, fmt.Errorf("missing trojan settings")
		}

		// Convert trojanSettings to TrojanConfig
		var trojanConfig common.TrojanConfig
		var trojanSettingsBytes []byte

		switch v := trojanSettings.(type) {
		case string:
			trojanSettingsBytes = []byte(v)
		case map[string]interface{}:
			var err error
			trojanSettingsBytes, err = json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal trojan settings: %v", err)
			}
		default:
			return nil, fmt.Errorf("invalid trojan settings type: %T", trojanSettings)
		}

		if err := json.Unmarshal(trojanSettingsBytes, &trojanConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal trojan config: %v", err)
		}

		server, err = NewTrojanServer(logger, proxyServer)
	case "shadowsocks":
		server, err = NewShadowsocksServer(logger, &common.ShadowsocksConfig{
			Password: settings["password"].(string),
			Method:   settings["method"].(string),
		}, proxyServer)
	default:
		return nil, fmt.Errorf("unsupported protocol type: %s", proxy.Type)
	}

	if err != nil {
		return nil, err
	}

	return server, nil
}

// StartProxy starts a proxy service
func (s *Service) StartProxy(proxy *common.ProxyConfig) error {
	s.Lock()
	defer s.Unlock()

	// Check if port is already in use
	for _, srv := range s.servers {
		if srv.GetPort() == proxy.Port {
			return fmt.Errorf("port %d is already in use", proxy.Port)
		}
	}

	// Create proxy server
	server, err := NewProxyServer(s.logger, proxy)
	if err != nil {
		return fmt.Errorf("failed to create proxy server: %v", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start proxy server: %v", err)
	}

	// Store server
	s.servers[proxy.ID] = server

	return nil
}

// StopProxy stops a proxy service
func (s *Service) StopProxy(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, ok := s.servers[id]
	if !ok {
		return fmt.Errorf("no proxy server found with ID %d", id)
	}

	// Stop server
	if err := server.Stop(); err != nil {
		return fmt.Errorf("failed to stop proxy server: %v", err)
	}

	// Remove server from map
	delete(s.servers, id)

	return nil
}

// LoadProxies loads all proxies from the database
func (s *Service) LoadProxies(proxies []*common.ProxyConfig) error {
	s.Lock()
	defer s.Unlock()

	for _, proxy := range proxies {
		if !proxy.Enabled {
			continue
		}

		server, err := NewProxyServer(s.logger, proxy)
		if err != nil {
			log.Printf("Failed to create proxy server: %v", err)
			continue
		}

		if err := server.Start(); err != nil {
			log.Printf("Failed to start proxy server: %v", err)
			continue
		}

		s.servers[proxy.ID] = server
	}

	return nil
}

// Create creates a new proxy
func (s *Service) Create(proxy *common.ProxyConfig) error {
	// Validate proxy
	if proxy.Port <= 0 || proxy.Port > 65535 {
		return fmt.Errorf("invalid port: %d", proxy.Port)
	}

	// Start proxy
	if err := s.StartProxy(proxy); err != nil {
		return err
	}

	return nil
}

// Update updates a proxy
func (s *Service) Update(proxy *common.ProxyConfig) error {
	// Stop existing proxy if any
	s.StopProxy(proxy.ID)

	// Start new proxy if enabled
	if proxy.Enabled {
		if err := s.StartProxy(proxy); err != nil {
			return err
		}
	}

	return nil
}

// Delete deletes a proxy
func (s *Service) Delete(id int64) error {
	return s.StopProxy(id)
}

// List returns all proxies
func (s *Service) List() []*model.Proxy {
	s.RLock()
	defer s.RUnlock()

	proxies := make([]*model.Proxy, 0, len(s.servers))
	for id, server := range s.servers {
		proxy := &model.Proxy{
			Base: model.Base{
				ID: id,
			},
			Upload:   server.GetUpload(),
			Download: server.GetDownload(),
		}
		proxies = append(proxies, proxy)
	}

	return proxies
}

// UpdateTraffic updates traffic statistics for a proxy
func (s *Service) UpdateTraffic(id int64, upload, download int64) error {
	s.Lock()
	defer s.Unlock()

	server, ok := s.servers[id]
	if !ok {
		return fmt.Errorf("proxy not found: %d", id)
	}

	server.UpdateTraffic(upload, download)

	// Use the database function that matches the actual expected type
	if err := database.DBInstance.UpdateTraffic(uint(id), server.GetUpload(), server.GetDownload()); err != nil {
		return fmt.Errorf("failed to update traffic stats: %v", err)
	}

	return nil
}

// Enable enables a proxy
func (s *Service) Enable(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, ok := s.servers[id]
	if !ok {
		return fmt.Errorf("proxy not found: %d", id)
	}

	// If already running, nothing to do
	if server != nil {
		return nil
	}

	// Get proxy from database
	proxy, err := database.DBInstance.GetProxyByID(id)
	if err != nil {
		return fmt.Errorf("failed to get proxy: %v", err)
	}

	// Enable the proxy in the database
	if err := database.DBInstance.Enable(uint(id)); err != nil {
		return fmt.Errorf("failed to enable proxy in database: %v", err)
	}

	// Create a ProxyConfig from the database proxy
	proxyConfig := &common.ProxyConfig{
		ID:       id,
		UserID:   proxy.UserID,
		Port:     proxy.Port,
		Settings: proxy.Settings,
		Enabled:  true,
		Type:     proxy.Protocol,
	}

	// Start proxy
	if err := s.StartProxy(proxyConfig); err != nil {
		return fmt.Errorf("failed to start proxy: %v", err)
	}

	return nil
}

// Disable disables a proxy
func (s *Service) Disable(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, ok := s.servers[id]
	if !ok {
		return fmt.Errorf("proxy not found: %d", id)
	}

	// Stop proxy
	if err := server.Stop(); err != nil {
		return fmt.Errorf("failed to stop proxy: %v", err)
	}

	// Remove from map
	delete(s.servers, id)

	// Update database
	if err := database.DBInstance.Disable(uint(id)); err != nil {
		return fmt.Errorf("failed to disable proxy in database: %v", err)
	}

	return nil
}

// UpdateLastActive updates the last active time for a proxy
func (s *Service) UpdateLastActive(id int64) error {
	s.Lock()
	defer s.Unlock()

	server, ok := s.servers[id]
	if !ok {
		return fmt.Errorf("proxy not found: %d", id)
	}

	// Update last active time
	now := time.Now()
	server.UpdateLastActive(now)

	// Update database
	if err := database.DBInstance.UpdateLastActive(uint(id)); err != nil {
		return fmt.Errorf("failed to update proxy last active time: %v", err)
	}

	return nil
}

// ProxyServerManager proxy server manager
type ProxyServerManager struct {
	servers map[int64]*common.ProxyInstance
	mu      sync.RWMutex
}

// NewProxyServerManager creates a new proxy server manager
func NewProxyServerManager() *ProxyServerManager {
	return &ProxyServerManager{
		servers: make(map[int64]*common.ProxyInstance),
	}
}

// AddServer adds a proxy server
func (m *ProxyServerManager) AddServer(proxy *common.ProxyInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[proxy.ID]; exists {
		return fmt.Errorf("server with ID %d already exists", proxy.ID)
	}

	m.servers[proxy.ID] = proxy
	return nil
}

// RemoveServer removes a proxy server
func (m *ProxyServerManager) RemoveServer(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.servers[id]; !exists {
		return fmt.Errorf("server with ID %d does not exist", id)
	}

	delete(m.servers, id)
	return nil
}

// GetServer gets a proxy server
func (m *ProxyServerManager) GetServer(id int64) (*common.ProxyInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[id]
	if !exists {
		return nil, fmt.Errorf("server with ID %d does not exist", id)
	}

	return server, nil
}

// GetAllServers gets all proxy servers
func (m *ProxyServerManager) GetAllServers() []*common.ProxyInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]*common.ProxyInstance, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}

	return servers
}

// StartServer starts a proxy server
func (m *ProxyServerManager) StartServer(id int64) error {
	server, err := m.GetServer(id)
	if err != nil {
		return err
	}

	return server.Server.Start()
}

// StopServer stops a proxy server
func (m *ProxyServerManager) StopServer(id int64) error {
	server, err := m.GetServer(id)
	if err != nil {
		return err
	}

	return server.Server.Stop()
}

// GetServerStats gets proxy server stats
func (m *ProxyServerManager) GetServerStats(id int64) (*common.ProxyStats, error) {
	server, err := m.GetServer(id)
	if err != nil {
		return nil, err
	}

	return &common.ProxyStats{
		Upload:    server.Upload,
		Download:  server.Download,
		Timestamp: time.Now().Unix(),
	}, nil
}

// UpdateServerStats updates proxy server stats
func (m *ProxyServerManager) UpdateServerStats(id int64, upload, download int64) error {
	server, err := m.GetServer(id)
	if err != nil {
		return err
	}

	server.Upload += upload
	server.Download += download
	server.LastActiveAt = time.Now()

	return nil
}

// GetProxyByID gets a proxy by ID
func (s *Service) GetProxyByID(id int64) (*common.Proxy, error) {
	return database.DBInstance.GetProxyByID(id)
}

// GetProxyByPort gets a proxy by port
func (s *Service) GetProxyByPort(port int) (*common.Proxy, error) {
	proxies, err := database.DBInstance.GetAllProxies()
	if err != nil {
		return nil, err
	}

	for _, proxy := range proxies {
		if proxy.Port == port {
			return proxy, nil
		}
	}

	return nil, fmt.Errorf("no proxy found with port %d", port)
}

// ListUserProxies lists all proxies for a user
func (s *Service) ListUserProxies(userID int64) ([]*common.Proxy, error) {
	return database.DBInstance.GetUserProxies(userID)
}
