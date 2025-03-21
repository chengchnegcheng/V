package model

// 使用 model/proxy.go 中已定义的 ProxyProtocol 类型
type ProtocolType string

const (
	ProtocolVMess       ProtocolType = "vmess"
	ProtocolVLESS       ProtocolType = "vless"
	ProtocolTrojan      ProtocolType = "trojan"
	ProtocolShadowsocks ProtocolType = "shadowsocks"
	ProtocolDokodemo    ProtocolType = "dokodemo-door"
	ProtocolSocks       ProtocolType = "socks"
	ProtocolHTTP        ProtocolType = "http"
)

// VMessSettings VMess 协议配置
type VMessSettings struct {
	UUID          string `json:"uuid"`
	AlterID       int    `json:"alterId"`
	Security      string `json:"security"`
	Network       string `json:"network"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	TLS           bool   `json:"tls"`
	AllowInsecure bool   `json:"allowInsecure"`
}

// VLESSSettings VLESS 协议配置
type VLESSSettings struct {
	UUID          string `json:"uuid"`
	Flow          string `json:"flow"`
	Network       string `json:"network"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	TLS           bool   `json:"tls"`
	AllowInsecure bool   `json:"allowInsecure"`
}

// TrojanSettings Trojan 协议配置
type TrojanSettings struct {
	Password string `json:"password"`
	Network  string `json:"network"`
	Host     string `json:"host"`
	Path     string `json:"path"`
	TLS      bool   `json:"tls"`
}

// ShadowsocksSettings Shadowsocks 协议配置
type ShadowsocksSettings struct {
	Method        string `json:"method"`
	Password      string `json:"password"`
	Network       string `json:"network"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Path          string `json:"path"`
	Plugin        string `json:"plugin,omitempty"`
	PluginOpts    string `json:"plugin_opts,omitempty"`
	AllowInsecure bool   `json:"allow_insecure"`
}

// DokodemoSettings Dokodemo-door 协议配置
type DokodemoSettings struct {
	Network        string `json:"network"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Timeout        int    `json:"timeout"`
	FollowRedirect bool   `json:"follow_redirect"`
}

// SocksSettings Socks 协议配置
type SocksSettings struct {
	Auth          string `json:"auth"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Network       string `json:"network"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	UDP           bool   `json:"udp"`
	AllowInsecure bool   `json:"allow_insecure"`
}

// HTTPSettings HTTP 协议配置
type HTTPSettings struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Network       string `json:"network"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Path          string `json:"path"`
	TLS           bool   `json:"tls"`
	AllowInsecure bool   `json:"allow_insecure"`
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
	CertFile      string `json:"cert_file,omitempty"`
	KeyFile       string `json:"key_file,omitempty"`
	ServerName    string `json:"server_name,omitempty"`
	AllowInsecure bool   `json:"allow_insecure,omitempty"`
}

// WSSettings WebSocket设置
type WSSettings struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}
