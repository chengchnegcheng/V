package proxy

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"v/common"
	"v/logger"
	"v/model"
)

// Manager 代理管理器
type Manager struct {
	logger  *logger.Logger
	db      model.DB
	proxies map[int64]*common.Proxy
	servers map[int64]common.Server
	mu      sync.RWMutex
}

// New 创建代理管理器
func New(logger *logger.Logger, db model.DB) *Manager {
	return &Manager{
		logger:  logger,
		db:      db,
		proxies: make(map[int64]*common.Proxy),
		servers: make(map[int64]common.Server),
	}
}

// Create 创建代理
func (m *Manager) Create(userID int64, port int, protocol common.ProtocolType, config *common.Config) (*common.Proxy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查端口是否被占用
	if m.checkPort(port) {
		return nil, fmt.Errorf("port %d is already in use", port)
	}

	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	// 创建代理
	proxy := &common.Proxy{
		UserID:   userID,
		Port:     port,
		Protocol: string(protocol),
		Config:   string(configJSON),
		Enabled:  true,
	}

	// 保存到数据库
	if err := m.db.CreateProxy(proxy); err != nil {
		return nil, err
	}

	// 创建服务器
	server, err := m.createServer(proxy)
	if err != nil {
		return nil, err
	}

	// 保存到内存
	m.proxies[proxy.ID] = proxy
	m.servers[proxy.ID] = server

	return proxy, nil
}

// Get 获取代理
func (m *Manager) Get(id int64) (*common.Proxy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proxy, exists := m.proxies[id]
	if !exists {
		return nil, fmt.Errorf("proxy %d not found", id)
	}

	return proxy, nil
}

// Update 更新代理
func (m *Manager) Update(id int64, config *common.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	proxy, exists := m.proxies[id]
	if !exists {
		return fmt.Errorf("proxy %d not found", id)
	}

	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// 更新配置
	proxy.Config = string(configJSON)

	// 保存到数据库
	if err := m.db.UpdateProxy(proxy); err != nil {
		return err
	}

	// 重启服务器
	if server, exists := m.servers[id]; exists {
		server.Stop()
		delete(m.servers, id)
	}

	server, err := m.createServer(proxy)
	if err != nil {
		return err
	}

	m.servers[id] = server
	return nil
}

// Delete 删除代理
func (m *Manager) Delete(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.proxies[id]
	if !exists {
		return fmt.Errorf("proxy %d not found", id)
	}

	// 停止服务器
	if server, exists := m.servers[id]; exists {
		server.Stop()
		delete(m.servers, id)
	}

	// 从数据库删除
	if err := m.db.DeleteProxy(id); err != nil {
		return err
	}

	delete(m.proxies, id)
	return nil
}

// Start 启动所有代理
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, proxy := range m.proxies {
		if !proxy.Enabled {
			continue
		}

		server, err := m.createServer(proxy)
		if err != nil {
			return err
		}

		m.servers[proxy.ID] = server
	}

	return nil
}

// Stop 停止所有代理
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, server := range m.servers {
		if err := server.Stop(); err != nil {
			return err
		}
	}

	m.servers = make(map[int64]common.Server)
	return nil
}

// UpdateTraffic 更新流量统计
func (m *Manager) UpdateTraffic(id int64, up, down int64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proxy, exists := m.proxies[id]
	if !exists {
		return fmt.Errorf("proxy %d not found", id)
	}

	proxy.LastActiveAt = time.Now()
	proxy.Upload += up
	proxy.Download += down

	return m.db.UpdateProxy(proxy)
}

// checkPort 检查端口是否被占用
func (m *Manager) checkPort(port int) bool {
	for _, proxy := range m.proxies {
		if proxy.Port == port {
			return true
		}
	}
	return false
}

// createServer 创建代理服务器
func (m *Manager) createServer(proxy *common.Proxy) (common.Server, error) {
	switch common.ProtocolType(proxy.Protocol) {
	case common.ProtocolVMess:
		return NewVMessServer(m.logger, proxy)
	case common.ProtocolVLESS:
		return NewVLESServer(m.logger, proxy)
	case common.ProtocolTrojan:
		return NewTrojanServer(m.logger, proxy)
	case common.ProtocolShadowsocks:
		return NewShadowsocksServer(m.logger, proxy)
	case common.ProtocolDokodemo:
		return NewDokodemoServer(m.logger, proxy)
	case common.ProtocolSocks:
		return NewSocksServer(m.logger, proxy)
	case common.ProtocolHTTP:
		return NewHTTPServer(m.logger, proxy)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proxy.Protocol)
	}
}
