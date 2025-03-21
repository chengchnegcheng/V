package protocol

import (
	"v/logger"
	"v/model"
	"v/settings"
)

// Manager 协议管理器
type Manager struct {
	log      *logger.Logger
	settings *settings.Manager
	db       model.DB
}

// New 创建协议管理器
func New(log *logger.Logger, settings *settings.Manager, db model.DB) *Manager {
	return &Manager{
		log:      log,
		settings: settings,
		db:       db,
	}
}

// ListProtocols 列出所有协议
func (m *Manager) ListProtocols(page, pageSize int) ([]*model.Protocol, error) {
	return m.db.ListProtocols(page, pageSize)
}

// GetTotalProtocols 获取协议总数
func (m *Manager) GetTotalProtocols() (int64, error) {
	return m.db.GetTotalProtocols()
}

// GetProtocol 获取指定协议
func (m *Manager) GetProtocol(id int64) (*model.Protocol, error) {
	return m.db.GetProtocol(id)
}

// CreateProtocol 创建协议
func (m *Manager) CreateProtocol(protocol *model.Protocol) error {
	return m.db.CreateProtocol(protocol)
}

// UpdateProtocol 更新协议
func (m *Manager) UpdateProtocol(protocol *model.Protocol) error {
	return m.db.UpdateProtocol(protocol)
}

// DeleteProtocol 删除协议
func (m *Manager) DeleteProtocol(id int64) error {
	return m.db.DeleteProtocol(id)
}

// GetProtocolStats 获取协议统计
func (m *Manager) GetProtocolStats() ([]*model.ProtocolStats, error) {
	// 实际应从数据库获取所有协议的统计数据
	return nil, nil
}

// GetSupportedProtocolTypes 获取支持的协议类型
func (m *Manager) GetSupportedProtocolTypes() []string {
	return []string{
		"vmess",
		"vless",
		"trojan",
		"shadowsocks",
		"socks",
		"http",
	}
}
