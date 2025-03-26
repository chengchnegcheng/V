package protocol

import (
	"encoding/json"
	"errors"
	"path/filepath"

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
	Log       XrayLogConfig     `json:"log"`
	Inbounds  []XrayInbound     `json:"inbounds"`
	Outbounds []XrayOutbound    `json:"outbounds"`
	Routing   XrayRoutingConfig `json:"routing"`
	DNS       *XrayDNSConfig    `json:"dns,omitempty"`
	Policy    *XrayPolicyConfig `json:"policy,omitempty"`
}

// XrayInbound Xray 入站配置
type XrayInbound struct {
	Port           int                 `json:"port"`
	Protocol       string              `json:"protocol"`
	Listen         string              `json:"listen,omitempty"`
	Settings       interface{}         `json:"settings"`
	StreamSettings *XrayStreamSettings `json:"streamSettings,omitempty"`
	Tag            string              `json:"tag,omitempty"`
	Sniffing       *XraySniffingConfig `json:"sniffing,omitempty"`
}

// XrayOutbound Xray 出站配置
type XrayOutbound struct {
	Protocol       string              `json:"protocol"`
	Settings       interface{}         `json:"settings"`
	StreamSettings *XrayStreamSettings `json:"streamSettings,omitempty"`
	Tag            string              `json:"tag,omitempty"`
	ProxySettings  *XrayProxySettings  `json:"proxySettings,omitempty"`
	Mux            *XrayMuxConfig      `json:"mux,omitempty"`
}

// XrayStreamSettings Xray 传输配置
type XrayStreamSettings struct {
	Network  string             `json:"network"`
	Security string             `json:"security,omitempty"`
	TLS      *XrayTLSConfig     `json:"tls,omitempty"`
	Reality  *XrayRealityConfig `json:"reality,omitempty"`
	WS       *XrayWSConfig      `json:"wsSettings,omitempty"`
	HTTP     *XrayHTTPConfig    `json:"httpSettings,omitempty"`
	QUIC     *XrayQUICConfig    `json:"quicSettings,omitempty"`
	GRPC     *XrayGRPCConfig    `json:"grpcSettings,omitempty"`
	TCP      *XrayTCPConfig     `json:"tcpSettings,omitempty"`
	Sockopt  *XraySockoptConfig `json:"sockopt,omitempty"`
}

// XrayTLSConfig Xray TLS 配置
type XrayTLSConfig struct {
	ServerName    string                  `json:"serverName"`
	AllowInsecure bool                    `json:"allowInsecure,omitempty"`
	Fingerprint   string                  `json:"fingerprint,omitempty"`
	Alpn          []string                `json:"alpn,omitempty"`
	Certificates  []XrayCertificateConfig `json:"certificates,omitempty"`
}

// XrayCertificateConfig Xray 证书配置
type XrayCertificateConfig struct {
	CertificateFile string `json:"certificateFile,omitempty"`
	KeyFile         string `json:"keyFile,omitempty"`
	Certificate     string `json:"certificate,omitempty"`
	Key             string `json:"key,omitempty"`
}

// XrayRealityConfig Xray Reality 配置
type XrayRealityConfig struct {
	Show         bool     `json:"show"`
	Dest         string   `json:"dest"`
	Xver         int      `json:"xver"`
	ServerNames  []string `json:"serverNames"`
	PrivateKey   string   `json:"privateKey"`
	MinClientVer string   `json:"minClientVer,omitempty"`
	MaxClientVer string   `json:"maxClientVer,omitempty"`
	MaxTimeDiff  int      `json:"maxTimeDiff,omitempty"`
	ShortIds     []string `json:"shortIds"`
}

// XrayWSConfig Xray WebSocket 配置
type XrayWSConfig struct {
	Path                string            `json:"path"`
	Headers             map[string]string `json:"headers,omitempty"`
	MaxEarlyData        int               `json:"maxEarlyData,omitempty"`
	EarlyDataHeaderName string            `json:"earlyDataHeaderName,omitempty"`
}

// XrayHTTPConfig Xray HTTP/2 配置
type XrayHTTPConfig struct {
	Path               string              `json:"path"`
	Host               []string            `json:"host,omitempty"`
	ReadIdleTimeout    int                 `json:"read_idle_timeout,omitempty"`
	HealthCheckTimeout int                 `json:"health_check_timeout,omitempty"`
	Headers            map[string][]string `json:"headers,omitempty"`
}

// XrayQUICConfig Xray QUIC 配置
type XrayQUICConfig struct {
	Security string `json:"security"`
	Key      string `json:"key"`
	Header   struct {
		Type string `json:"type"`
	} `json:"header"`
}

// XrayGRPCConfig Xray gRPC 配置
type XrayGRPCConfig struct {
	ServiceName        string `json:"serviceName"`
	MultiMode          bool   `json:"multiMode,omitempty"`
	IdleTimeout        int    `json:"idle_timeout,omitempty"`
	InitialWindowsSize int    `json:"initial_windows_size,omitempty"`
}

// XrayTCPConfig Xray TCP 配置
type XrayTCPConfig struct {
	AcceptProxyProtocol bool `json:"acceptProxyProtocol,omitempty"`
	Header              struct {
		Type     string                 `json:"type"`
		Request  map[string]interface{} `json:"request,omitempty"`
		Response map[string]interface{} `json:"response,omitempty"`
	} `json:"header,omitempty"`
}

// XraySockoptConfig Xray Sockopt 配置
type XraySockoptConfig struct {
	Mark                 int    `json:"mark,omitempty"`
	TCPFastOpen          bool   `json:"tcpFastOpen,omitempty"`
	Tproxy               string `json:"tproxy,omitempty"`
	DomainStrategy       string `json:"domainStrategy,omitempty"`
	DialerProxy          string `json:"dialerProxy,omitempty"`
	TCPKeepAliveInterval int    `json:"tcpKeepAliveInterval,omitempty"`
}

// XraySniffingConfig Xray 流量嗅探配置
type XraySniffingConfig struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
	MetadataOnly bool     `json:"metadataOnly,omitempty"`
}

// XrayRoutingConfig Xray 路由配置
type XrayRoutingConfig struct {
	DomainStrategy string            `json:"domainStrategy"`
	Rules          []XrayRoutingRule `json:"rules"`
	Balancers      []XrayBalancer    `json:"balancers,omitempty"`
}

// XrayRoutingRule Xray 路由规则
type XrayRoutingRule struct {
	Type        string   `json:"type"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	SourcePort  string   `json:"sourcePort,omitempty"`
	Network     string   `json:"network,omitempty"`
	Source      []string `json:"source,omitempty"`
	User        []string `json:"user,omitempty"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
	Attrs       string   `json:"attrs,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	BalancerTag string   `json:"balancerTag,omitempty"`
}

// XrayBalancer Xray 负载均衡器
type XrayBalancer struct {
	Tag      string   `json:"tag"`
	Selector []string `json:"selector"`
	Strategy struct {
		Type string `json:"type"`
	} `json:"strategy,omitempty"`
}

// XrayLogConfig Xray 日志配置
type XrayLogConfig struct {
	Access   string `json:"access,omitempty"`
	Error    string `json:"error,omitempty"`
	Loglevel string `json:"loglevel,omitempty"`
}

// XrayDNSConfig Xray DNS 配置
type XrayDNSConfig struct {
	Servers  []interface{}     `json:"servers"`
	Hosts    map[string]string `json:"hosts,omitempty"`
	ClientIP string            `json:"clientIp,omitempty"`
	Tag      string            `json:"tag,omitempty"`
}

// XrayPolicyConfig Xray 策略配置
type XrayPolicyConfig struct {
	Levels map[string]XrayLevelPolicyConfig `json:"levels,omitempty"`
	System *XraySystemPolicyConfig          `json:"system,omitempty"`
}

// XrayLevelPolicyConfig Xray 等级策略配置
type XrayLevelPolicyConfig struct {
	HandshakeTime     int  `json:"handshake,omitempty"`
	ConnIdle          int  `json:"connIdle,omitempty"`
	UplinkOnly        int  `json:"uplinkOnly,omitempty"`
	DownlinkOnly      int  `json:"downlinkOnly,omitempty"`
	StatsUserUplink   bool `json:"statsUserUplink,omitempty"`
	StatsUserDownlink bool `json:"statsUserDownlink,omitempty"`
	BufferSize        int  `json:"bufferSize,omitempty"`
}

// XraySystemPolicyConfig Xray 系统策略配置
type XraySystemPolicyConfig struct {
	StatsInboundUplink    bool `json:"statsInboundUplink,omitempty"`
	StatsInboundDownlink  bool `json:"statsInboundDownlink,omitempty"`
	StatsOutboundUplink   bool `json:"statsOutboundUplink,omitempty"`
	StatsOutboundDownlink bool `json:"statsOutboundDownlink,omitempty"`
}

// XrayProxySettings Xray 代理设置
type XrayProxySettings struct {
	Tag string `json:"tag"`
}

// XrayMuxConfig Xray 多路复用配置
type XrayMuxConfig struct {
	Enabled     bool `json:"enabled"`
	Concurrency int  `json:"concurrency,omitempty"`
}

// XrayFreedomSettings Xray freedom设置
type XrayFreedomSettings struct {
	DomainStrategy string `json:"domainStrategy,omitempty"`
	Redirect       string `json:"redirect,omitempty"`
	UserLevel      int    `json:"userLevel,omitempty"`
}

// GenerateXrayConfig 生成 Xray 配置
func (m *ProtocolManager) GenerateXrayConfig(protocol *model.Protocol) (*XrayConfig, error) {
	config := &XrayConfig{
		Log: XrayLogConfig{
			Access:   "none",
			Error:    filepath.Join("logs", "xray.log"),
			Loglevel: "warning",
		},
		Inbounds:  make([]XrayInbound, 0),
		Outbounds: make([]XrayOutbound, 0),
		Routing: XrayRoutingConfig{
			DomainStrategy: "AsIs",
			Rules: []XrayRoutingRule{
				{
					Type:        "field",
					InboundTag:  []string{"api"},
					OutboundTag: "api",
				},
			},
		},
	}

	// 根据协议类型生成相应配置
	var settings interface{}
	var err error

	switch protocol.Type {
	case "vmess":
		settings, err = m.GenerateVMessConfig(protocol)
		if err == nil {
			// 解析 VMess 配置
			vmessSettings, ok := settings.(*model.VMessSettings)
			if ok {
				// 创建流设置
				streamSettings := &XrayStreamSettings{
					Network: vmessSettings.Network,
				}

				// 检查并设置 TLS
				if vmessSettings.TLS {
					streamSettings.Security = "tls"
					streamSettings.TLS = &XrayTLSConfig{
						ServerName:    vmessSettings.Host,
						AllowInsecure: vmessSettings.AllowInsecure,
					}
				}

				// 根据网络类型设置特定配置
				switch vmessSettings.Network {
				case "ws":
					streamSettings.WS = &XrayWSConfig{
						Path: vmessSettings.Path,
						Headers: map[string]string{
							"Host": vmessSettings.Host,
						},
					}
				case "http":
					streamSettings.HTTP = &XrayHTTPConfig{
						Path: vmessSettings.Path,
						Host: []string{vmessSettings.Host},
					}
				}

				// 配置入站
				config.Inbounds = append(config.Inbounds, XrayInbound{
					Port:           protocol.Port,
					Protocol:       protocol.Type,
					Settings:       settings,
					StreamSettings: streamSettings,
					Sniffing: &XraySniffingConfig{
						Enabled:      true,
						DestOverride: []string{"http", "tls"},
					},
				})
			}
		}
	case "vless":
		settings, err = m.GenerateVLESSConfig(protocol)
		if err == nil {
			// 解析 VLESS 配置
			vlessSettings, ok := settings.(*model.VLESSSettings)
			if ok {
				// 创建流设置
				streamSettings := &XrayStreamSettings{
					Network: vlessSettings.Network,
				}

				// 检查并设置 TLS
				if vlessSettings.TLS {
					streamSettings.Security = "tls"
					streamSettings.TLS = &XrayTLSConfig{
						ServerName:    vlessSettings.Host,
						AllowInsecure: vlessSettings.AllowInsecure,
					}
				}

				// 根据网络类型设置特定配置
				switch vlessSettings.Network {
				case "ws":
					streamSettings.WS = &XrayWSConfig{
						Path: vlessSettings.Path,
						Headers: map[string]string{
							"Host": vlessSettings.Host,
						},
					}
				case "http":
					streamSettings.HTTP = &XrayHTTPConfig{
						Path: vlessSettings.Path,
						Host: []string{vlessSettings.Host},
					}
				case "grpc":
					streamSettings.GRPC = &XrayGRPCConfig{
						ServiceName: vlessSettings.Path,
					}
				}

				// 配置入站
				config.Inbounds = append(config.Inbounds, XrayInbound{
					Port:           protocol.Port,
					Protocol:       protocol.Type,
					Settings:       settings,
					StreamSettings: streamSettings,
					Sniffing: &XraySniffingConfig{
						Enabled:      true,
						DestOverride: []string{"http", "tls"},
					},
				})
			}
		}
	case "trojan":
		settings, err = m.GenerateTrojanConfig(protocol)
		if err == nil {
			// 解析 Trojan 配置
			trojanSettings, ok := settings.(*model.TrojanSettings)
			if ok {
				// 创建流设置
				streamSettings := &XrayStreamSettings{
					Network: trojanSettings.Network,
				}

				// 对于 Trojan 默认启用 TLS
				streamSettings.Security = "tls"

				// 设置 SNI，优先使用 SNI 字段，如果为空则使用 Host 字段
				var serverName string
				if trojanSettings.SNI != "" {
					serverName = trojanSettings.SNI
				} else {
					serverName = trojanSettings.Host
				}

				streamSettings.TLS = &XrayTLSConfig{
					ServerName: serverName,
				}

				// 根据网络类型设置特定配置
				switch trojanSettings.Network {
				case "ws":
					streamSettings.WS = &XrayWSConfig{
						Path: trojanSettings.Path,
						Headers: map[string]string{
							"Host": trojanSettings.Host,
						},
					}
				case "grpc":
					streamSettings.GRPC = &XrayGRPCConfig{
						ServiceName: trojanSettings.Path,
					}
				}

				// 配置入站
				config.Inbounds = append(config.Inbounds, XrayInbound{
					Port:           protocol.Port,
					Protocol:       protocol.Type,
					Settings:       settings,
					StreamSettings: streamSettings,
					Sniffing: &XraySniffingConfig{
						Enabled:      true,
						DestOverride: []string{"http", "tls"},
					},
				})
			}
		}
	case "shadowsocks":
		settings, err = m.GenerateShadowsocksConfig(protocol)
		if err == nil {
			// 解析 Shadowsocks 配置
			ssSettings, ok := settings.(*model.ShadowsocksSettings)
			if ok {
				// 创建流设置
				streamSettings := &XrayStreamSettings{
					Network: ssSettings.Network,
				}

				// 根据网络类型设置特定配置
				switch ssSettings.Network {
				case "ws":
					streamSettings.WS = &XrayWSConfig{
						Path: ssSettings.Path,
						Headers: map[string]string{
							"Host": ssSettings.Host,
						},
					}
				}

				// 配置入站
				config.Inbounds = append(config.Inbounds, XrayInbound{
					Port:           protocol.Port,
					Protocol:       protocol.Type,
					Settings:       settings,
					StreamSettings: streamSettings,
					Sniffing: &XraySniffingConfig{
						Enabled:      true,
						DestOverride: []string{"http", "tls"},
					},
				})
			}
		}
	default:
		return nil, errors.New("unsupported protocol type")
	}

	if err != nil {
		return nil, err
	}

	// 如果没有成功添加入站配置，添加默认配置
	if len(config.Inbounds) == 0 {
		config.Inbounds = append(config.Inbounds, XrayInbound{
			Port:     protocol.Port,
			Protocol: protocol.Type,
			Settings: settings,
			Sniffing: &XraySniffingConfig{
				Enabled:      true,
				DestOverride: []string{"http", "tls"},
			},
		})
	}

	// 添加出站配置
	config.Outbounds = append(config.Outbounds, XrayOutbound{
		Protocol: "freedom",
		Tag:      "direct",
		Settings: XrayFreedomSettings{
			DomainStrategy: "UseIP",
		},
		Mux: &XrayMuxConfig{
			Enabled:     true,
			Concurrency: 8,
		},
	})

	// 添加黑洞出站
	config.Outbounds = append(config.Outbounds, XrayOutbound{
		Protocol: "blackhole",
		Tag:      "blocked",
		Settings: map[string]interface{}{},
	})

	return config, nil
}
