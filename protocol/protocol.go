package protocol

import (
	"encoding/json"
	"errors"

	"v/logger"
	"v/model"
	"v/settings"
)

// ProtocolManager 协议管理器
type ProtocolManager struct {
	logger   *logger.Logger
	settings *settings.Manager
	db       model.DB
}

// NewProtocolManager 创建协议管理器
func NewProtocolManager(logger *logger.Logger, settings *settings.Manager, db model.DB) *ProtocolManager {
	return &ProtocolManager{
		logger:   logger,
		settings: settings,
		db:       db,
	}
}

// Create 创建协议配置
func (m *ProtocolManager) Create(userID int64, protocolType model.ProtocolType, name string, port int, settings interface{}) (*model.Protocol, error) {
	// 验证端口是否可用
	if err := m.validatePort(port); err != nil {
		return nil, err
	}

	// 验证协议配置
	if err := m.ValidateProtocolSettings(protocolType, settings); err != nil {
		return nil, err
	}

	// 序列化设置
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	// 创建协议配置
	protocol := &model.Protocol{
		UserID:       userID,
		Type:         string(protocolType),
		Name:         name,
		Port:         port,
		Settings:     settingsJSON,
		Enable:       true,
		TrafficLimit: 0,
		TrafficUsed:  0,
	}

	// 保存到数据库
	if err := m.db.CreateProtocol(protocol); err != nil {
		return nil, err
	}

	return protocol, nil
}

// Get 获取协议配置
func (m *ProtocolManager) Get(id int64) (*model.Protocol, error) {
	return m.db.GetProtocol(id)
}

// GetByUserID 获取用户的协议配置列表
func (m *ProtocolManager) GetByUserID(userID int64) ([]*model.Protocol, error) {
	return m.db.GetProtocolsByUserID(userID)
}

// Update 更新协议配置
func (m *ProtocolManager) Update(protocol *model.Protocol) error {
	// 验证端口是否可用
	if err := m.validatePort(protocol.Port); err != nil {
		return err
	}

	return m.db.UpdateProtocol(protocol)
}

// Delete 删除协议配置
func (m *ProtocolManager) Delete(id int64) error {
	return m.db.DeleteProtocol(id)
}

// Enable 启用协议配置
func (m *ProtocolManager) Enable(id int64) error {
	protocol, err := m.Get(id)
	if err != nil {
		return err
	}

	protocol.Enable = true
	return m.Update(protocol)
}

// Disable 禁用协议配置
func (m *ProtocolManager) Disable(id int64) error {
	protocol, err := m.Get(id)
	if err != nil {
		return err
	}

	protocol.Enable = false
	return m.Update(protocol)
}

// UpdateTraffic 更新流量统计
func (m *ProtocolManager) UpdateTraffic(id int64, upload, download int64) error {
	protocol, err := m.Get(id)
	if err != nil {
		return err
	}

	protocol.TrafficUsed += upload + download
	return m.Update(protocol)
}

// validatePort 验证端口是否可用
func (m *ProtocolManager) validatePort(port int) error {
	if port < 1 || port > 65535 {
		return errors.New("invalid port number")
	}

	// 检查端口是否已被使用
	protocols, err := m.db.GetProtocolsByPort(port)
	if err != nil {
		return err
	}

	if len(protocols) > 0 {
		return errors.New("port already in use")
	}

	return nil
}

// GenerateVMessConfig 生成 VMess 配置
func (m *ProtocolManager) GenerateVMessConfig(protocol *model.Protocol) (*model.VMessSettings, error) {
	var settings model.VMessSettings
	if err := json.Unmarshal(protocol.Settings, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// GenerateVLESSConfig 生成 VLESS 配置
func (m *ProtocolManager) GenerateVLESSConfig(protocol *model.Protocol) (*model.VLESSSettings, error) {
	var settings model.VLESSSettings
	if err := json.Unmarshal(protocol.Settings, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// GenerateTrojanConfig 生成 Trojan 配置
func (m *ProtocolManager) GenerateTrojanConfig(protocol *model.Protocol) (*model.TrojanSettings, error) {
	var settings model.TrojanSettings
	if err := json.Unmarshal(protocol.Settings, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// GenerateShadowsocksConfig 生成 Shadowsocks 配置
func (m *ProtocolManager) GenerateShadowsocksConfig(protocol *model.Protocol) (*model.ShadowsocksSettings, error) {
	var settings model.ShadowsocksSettings
	if err := json.Unmarshal(protocol.Settings, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// ValidateVMessSettings 验证 VMess 配置
func (m *ProtocolManager) ValidateVMessSettings(settings *model.VMessSettings) error {
	if settings.UUID == "" {
		return errors.New("uuid is required")
	}
	if settings.AlterID < 0 {
		return errors.New("alterId must be non-negative")
	}
	if settings.Security == "" {
		settings.Security = "auto"
	}
	if settings.Network == "" {
		settings.Network = "tcp"
	}
	if settings.Host == "" {
		return errors.New("host is required")
	}
	return nil
}

// ValidateVLESSSettings 验证 VLESS 配置
func (m *ProtocolManager) ValidateVLESSSettings(settings *model.VLESSSettings) error {
	if settings.UUID == "" {
		return errors.New("uuid is required")
	}
	if settings.Network == "" {
		settings.Network = "tcp"
	}
	if settings.Host == "" {
		return errors.New("host is required")
	}
	return nil
}

// ValidateTrojanSettings 验证 Trojan 配置
func (m *ProtocolManager) ValidateTrojanSettings(settings *model.TrojanSettings) error {
	if settings.Password == "" {
		return errors.New("password is required")
	}
	if settings.Network == "" {
		settings.Network = "tcp"
	}
	if settings.Host == "" {
		return errors.New("host is required")
	}
	return nil
}

// ValidateShadowsocksSettings 验证 Shadowsocks 配置
func (m *ProtocolManager) ValidateShadowsocksSettings(settings *model.ShadowsocksSettings) error {
	if settings.Method == "" {
		return errors.New("method is required")
	}
	if settings.Password == "" {
		return errors.New("password is required")
	}
	if settings.Network == "" {
		settings.Network = "tcp"
	}
	if settings.Host == "" {
		return errors.New("host is required")
	}
	return nil
}

// ValidateProtocolSettings 验证协议配置
func (m *ProtocolManager) ValidateProtocolSettings(protocolType model.ProtocolType, settings interface{}) error {
	// 将model.ProtocolType转换为字符串进行比较
	protocolTypeStr := string(protocolType)

	switch protocolTypeStr {
	case "vmess":
		if vmessSettings, ok := settings.(*model.VMessSettings); ok {
			return m.ValidateVMessSettings(vmessSettings)
		}
		return errors.New("invalid VMess settings")
	case "vless":
		if vlessSettings, ok := settings.(*model.VLESSSettings); ok {
			return m.ValidateVLESSSettings(vlessSettings)
		}
		return errors.New("invalid VLESS settings")
	case "trojan":
		if trojanSettings, ok := settings.(*model.TrojanSettings); ok {
			return m.ValidateTrojanSettings(trojanSettings)
		}
		return errors.New("invalid Trojan settings")
	case "shadowsocks":
		if ssSettings, ok := settings.(*model.ShadowsocksSettings); ok {
			return m.ValidateShadowsocksSettings(ssSettings)
		}
		return errors.New("invalid Shadowsocks settings")
	default:
		return errors.New("unsupported protocol")
	}
}

// XrayConfig Xray 配置结构
type XrayConfig struct {
	Inbounds  []XrayInbound  `json:"inbounds"`
	Outbounds []XrayOutbound `json:"outbounds"`
}

// XrayInbound Xray 入站配置
type XrayInbound struct {
	Port           int                 `json:"port"`
	Protocol       string              `json:"protocol"`
	Settings       interface{}         `json:"settings"`
	StreamSettings *XrayStreamSettings `json:"streamSettings,omitempty"`
}

// XrayOutbound Xray 出站配置
type XrayOutbound struct {
	Protocol       string              `json:"protocol"`
	Settings       interface{}         `json:"settings"`
	StreamSettings *XrayStreamSettings `json:"streamSettings,omitempty"`
}

// XrayStreamSettings Xray 传输配置
type XrayStreamSettings struct {
	Network  string          `json:"network"`
	Security string          `json:"security,omitempty"`
	TLS      *XrayTLSConfig  `json:"tls,omitempty"`
	WS       *XrayWSConfig   `json:"ws,omitempty"`
	HTTP     *XrayHTTPConfig `json:"http,omitempty"`
}

// XrayTLSConfig Xray TLS 配置
type XrayTLSConfig struct {
	AllowInsecure bool   `json:"allowInsecure"`
	ServerName    string `json:"serverName"`
}

// XrayWSConfig Xray WebSocket 配置
type XrayWSConfig struct {
	Path string `json:"path"`
}

// XrayHTTPConfig Xray HTTP 配置
type XrayHTTPConfig struct {
	Path string `json:"path"`
}

// XrayFreedomSettings Xray freedom设置
type XrayFreedomSettings struct {
	DomainStrategy string `json:"domainStrategy,omitempty"`
}

// GenerateXrayConfig 生成 Xray 配置
func (m *ProtocolManager) GenerateXrayConfig(protocol *model.Protocol) (*XrayConfig, error) {
	config := &XrayConfig{
		Inbounds:  make([]XrayInbound, 0),
		Outbounds: make([]XrayOutbound, 0),
	}

	// 根据协议类型生成相应配置
	var settings interface{}
	var err error

	switch protocol.Type {
	case "vmess":
		settings, err = m.GenerateVMessConfig(protocol)
	case "vless":
		settings, err = m.GenerateVLESSConfig(protocol)
	case "trojan":
		settings, err = m.GenerateTrojanConfig(protocol)
	case "shadowsocks":
		settings, err = m.GenerateShadowsocksConfig(protocol)
	default:
		return nil, errors.New("unsupported protocol type")
	}

	if err != nil {
		return nil, err
	}

	// 添加入站配置
	config.Inbounds = append(config.Inbounds, XrayInbound{
		Port:     protocol.Port,
		Protocol: protocol.Type,
		Settings: settings,
	})

	// 添加出站配置
	config.Outbounds = append(config.Outbounds, XrayOutbound{
		Protocol: "freedom",
		Settings: XrayFreedomSettings{},
	})

	return config, nil
}
