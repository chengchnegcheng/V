package common

import (
	"io"
	"time"
)

// ProtocolType 协议类型
type ProtocolType string

// 协议类型常量
const (
	ProtocolVMess       ProtocolType = "vmess"
	ProtocolVLESS       ProtocolType = "vless"
	ProtocolTrojan      ProtocolType = "trojan"
	ProtocolShadowsocks ProtocolType = "shadowsocks"
	ProtocolDokodemo    ProtocolType = "dokodemo-door"
	ProtocolSocks       ProtocolType = "socks"
	ProtocolHTTP        ProtocolType = "http"
)

// Server 代理服务器接口
type Server interface {
	Start() error
	Stop() error
	GetPort() int
	GetProtocol() ProtocolType
	HandleConnection(io.ReadWriteCloser) error
}

// Config 代理配置
type Config struct {
	Protocol       ProtocolType    `json:"protocol"`
	Settings       interface{}     `json:"settings"`
	StreamSettings *StreamSettings `json:"stream_settings,omitempty"`
	Port           int             `json:"port"`
}

// StreamSettings 流设置
type StreamSettings struct {
	Network      string        `json:"network,omitempty"`
	Security     string        `json:"security,omitempty"`
	TLSSettings  *TLSSettings  `json:"tls,omitempty"`
	WSSettings   *WSSettings   `json:"ws,omitempty"`
	HTTPSettings *HTTPSettings `json:"http,omitempty"`
}

// TLSSettings TLS设置
type TLSSettings struct {
	CertFile   string `json:"cert_file,omitempty"`
	KeyFile    string `json:"key_file,omitempty"`
	ServerName string `json:"server_name,omitempty"`
}

// WSSettings WebSocket设置
type WSSettings struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// HTTPSettings HTTP设置
type HTTPSettings struct {
	Host    []string          `json:"host,omitempty"`
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Fields 日志字段
type Fields map[string]interface{}

// ProxyStats 代理统计
type ProxyStats struct {
	Upload    int64
	Download  int64
	Timestamp int64
}

// Proxy 代理配置
type Proxy struct {
	ID           int64      `json:"id" db:"id"`
	UserID       int64      `json:"user_id" db:"user_id"`
	Protocol     string     `json:"protocol" db:"protocol"`
	Port         int        `json:"port" db:"port"`
	Config       string     `json:"config" db:"config"`
	Settings     string     `json:"settings" db:"settings"`
	ListenAddr   string     `json:"listen_addr" db:"listen_addr"`
	RemoteAddr   string     `json:"remote_addr" db:"remote_addr"`
	Enabled      bool       `json:"enabled" db:"enabled"`
	Upload       int64      `json:"upload" db:"upload"`
	Download     int64      `json:"download" db:"download"`
	LastActiveAt time.Time  `json:"last_active_at" db:"last_active_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	ExpireAt     *time.Time `json:"expire_at" db:"expire_at"`
}

// TrafficStats 流量统计
type TrafficStats struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	ProxyID      int64     `json:"proxy_id" db:"proxy_id"`
	Upload       int64     `json:"upload" db:"upload"`
	Download     int64     `json:"download" db:"download"`
	Total        int64     `json:"total" db:"total"`
	TrafficLimit int64     `json:"traffic_limit" db:"traffic_limit"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// VMessConfig VMess 配置
type VMessConfig struct {
	ID       string `json:"id"`
	AlterID  int    `json:"alterId"`
	Security string `json:"security"`
}

// VLESSConfig VLESS 配置
type VLESSConfig struct {
	ID       string `json:"id"`
	Flow     string `json:"flow"`
	Security string `json:"security"`
}

// TrojanConfig Trojan 配置
type TrojanConfig struct {
	Password string `json:"password"`
	Security string `json:"security"`
	SSL      struct {
		CertFile string `json:"cert_file"`
		KeyFile  string `json:"key_file"`
	} `json:"ssl"`
}

// ShadowsocksConfig Shadowsocks 配置
type ShadowsocksConfig struct {
	Method     string `json:"method"`
	Password   string `json:"password"`
	Security   string `json:"security"`
	Plugin     string `json:"plugin,omitempty"`
	PluginOpts string `json:"plugin_opts,omitempty"`
}

// DokodemoConfig Dokodemo 配置
type DokodemoConfig struct {
	TargetAddr string `json:"target_addr"`
	TargetPort int    `json:"target_port"`
	Network    string `json:"network"`
	Timeout    int    `json:"timeout"`
}

// SocksConfig Socks 配置
type SocksConfig struct {
	Auth     string `json:"auth"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// HTTPConfig HTTP 配置
type HTTPConfig struct {
	Auth     string `json:"auth"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ProxyServerInfo 代理服务器
type ProxyServerInfo struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	Protocol     string                 `json:"protocol"`
	Port         int                    `json:"port"`
	Settings     map[string]interface{} `json:"settings"`
	Enabled      bool                   `json:"enabled"`
	Upload       int64                  `json:"upload"`
	Download     int64                  `json:"download"`
	LastActiveAt time.Time              `json:"last_active_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Server       Server                 `json:"-"`
}
