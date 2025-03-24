package common

import (
	"io"
)

// LegacyProxyServer represents legacy proxy server configuration (old format)
type LegacyProxyServer struct {
	Type     string                 `json:"type"`
	Port     int                    `json:"port"`
	Settings map[string]interface{} `json:"settings"`
}

// LegacyVMessConfig represents legacy VMess configuration (old format)
type LegacyVMessConfig struct {
	ID string `json:"id"`
}

// LegacyVLESSConfig represents legacy VLESS configuration (old format)
type LegacyVLESSConfig struct {
	ID string `json:"id"`
}

// LegacyTrojanConfig represents legacy Trojan configuration (old format)
type LegacyTrojanConfig struct {
	Password string `json:"password"`
}

// LegacyShadowsocksConfig represents legacy Shadowsocks configuration (old format)
type LegacyShadowsocksConfig struct {
	Password string `json:"password"`
	Method   string `json:"method"`
}

// LegacyServerInterface represents legacy server interface (old format)
type LegacyServerInterface interface {
	Start() error
	Stop() error
	HandleConnection(conn io.ReadWriteCloser) error
}

// LegacyTLSConfig represents legacy TLS configuration (old format)
type LegacyTLSConfig struct {
	CertFile   string `json:"cert_file"`
	KeyFile    string `json:"key_file"`
	ServerName string `json:"server_name"`
}
