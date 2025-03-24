package proxy

import (
	"encoding/json"
	"fmt"
	"io"
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
	// Convert common.Proxy to common.ProxyInstance
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(proxy.Settings), &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %v", err)
	}

	proxyInstance := &common.ProxyInstance{
		ID:       proxy.ID,
		UserID:   proxy.UserID,
		Type:     proxy.Protocol,
		Port:     proxy.Port,
		Settings: settings,
		Enabled:  proxy.Enabled,
	}

	var server common.ProxyServerInterface
	var err error

	switch common.ProtocolType(proxy.Protocol) {
	case common.ProtocolVMess:
		vmessConfig := &common.VMessConfig{}
		if vmessSettings, ok := settings["vmess"]; ok {
			vmessSettingsBytes, err := json.Marshal(vmessSettings)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal vmess settings: %v", err)
			}
			if err := json.Unmarshal(vmessSettingsBytes, vmessConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal vmess config: %v", err)
			}
		}
		server, err = NewVMessServer(m.logger, vmessConfig, proxyInstance)
	case common.ProtocolVLESS:
		vlessConfig := &common.VLESSConfig{}
		if vlessSettings, ok := settings["vless"]; ok {
			vlessSettingsBytes, err := json.Marshal(vlessSettings)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal vless settings: %v", err)
			}
			if err := json.Unmarshal(vlessSettingsBytes, vlessConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal vless config: %v", err)
			}
		}
		server, err = NewVLESSServer(m.logger, vlessConfig, proxyInstance)
	case common.ProtocolTrojan:
		server, err = NewTrojanServer(m.logger, proxyInstance)
	case common.ProtocolShadowsocks:
		ssConfig := &common.ShadowsocksConfig{}
		if ssSettings, ok := settings["shadowsocks"]; ok {
			ssSettingsBytes, err := json.Marshal(ssSettings)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal shadowsocks settings: %v", err)
			}
			if err := json.Unmarshal(ssSettingsBytes, ssConfig); err != nil {
				return nil, fmt.Errorf("failed to unmarshal shadowsocks config: %v", err)
			}
		}
		server, err = NewShadowsocksServer(m.logger, ssConfig, proxyInstance)
	case common.ProtocolDokodemo:
		// Implement conversion for other protocols
		return nil, fmt.Errorf("protocol not implemented yet: %s", proxy.Protocol)
	case common.ProtocolSocks:
		// Implement conversion for other protocols
		return nil, fmt.Errorf("protocol not implemented yet: %s", proxy.Protocol)
	case common.ProtocolHTTP:
		// Implement conversion for other protocols
		return nil, fmt.Errorf("protocol not implemented yet: %s", proxy.Protocol)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proxy.Protocol)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create server: %v", err)
	}

	// Adapt ProxyServerInterface to common.Server
	return &serverAdapter{server}, nil
}

// serverAdapter adapts ProxyServerInterface to common.Server
type serverAdapter struct {
	server common.ProxyServerInterface
}

// Start implements common.Server
func (a *serverAdapter) Start() error {
	return a.server.Start()
}

// Stop implements common.Server
func (a *serverAdapter) Stop() error {
	return a.server.Stop()
}

// GetPort implements common.Server
func (a *serverAdapter) GetPort() int {
	return a.server.GetPort()
}

// GetProtocol implements common.Server
func (a *serverAdapter) GetProtocol() common.ProtocolType {
	if getter, ok := a.server.(interface{ GetProtocol() common.ProtocolType }); ok {
		return getter.GetProtocol()
	}
	// Default fallback
	return common.ProtocolType("")
}

// HandleConnection implements common.Server
func (a *serverAdapter) HandleConnection(conn io.ReadWriteCloser) error {
	return a.server.HandleConnection(conn)
}
