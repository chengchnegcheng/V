package protocol

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"v/model"
)

// VMessLink VMess 链接结构
type VMessLink struct {
	V    string `json:"v"`
	PS   string `json:"ps"`
	Add  string `json:"add"`
	Port string `json:"port"`
	ID   string `json:"id"`
	Aid  int    `json:"aid"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	TLS  string `json:"tls"`
}

// VLESSLink VLESS 链接结构
type VLESSLink struct {
	ID         string `json:"id"`
	Flow       string `json:"flow"`
	Encryption string `json:"encryption"`
	FP         string `json:"fp"`
	Type       string `json:"type"`
	Host       string `json:"host"`
	Path       string `json:"path"`
	Port       string `json:"port"`
}

// TrojanLink Trojan 链接结构
type TrojanLink struct {
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Path     string `json:"path"`
}

// ShadowsocksLink Shadowsocks 链接结构
type ShadowsocksLink struct {
	Method   string `json:"method"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Path     string `json:"path"`
}

// GenerateVMessLink 生成 VMess 链接
func (m *ProtocolManager) GenerateVMessLink(protocol *model.Protocol) (string, error) {
	settings, err := m.GenerateVMessConfig(protocol)
	if err != nil {
		return "", err
	}

	link := VMessLink{
		V:    "2",
		PS:   protocol.Name,
		Add:  settings.Host,
		Port: fmt.Sprintf("%d", protocol.Port),
		ID:   settings.UUID,
		Aid:  settings.AlterID,
		Net:  settings.Network,
		Type: "none",
		Host: settings.Host,
		Path: settings.Path,
		TLS:  "none",
	}

	if settings.TLS {
		link.TLS = "tls"
	}

	jsonData, err := json.Marshal(link)
	if err != nil {
		return "", err
	}

	return "vmess://" + base64.StdEncoding.EncodeToString(jsonData), nil
}

// GenerateVLESSLink 生成 VLESS 链接
func (m *ProtocolManager) GenerateVLESSLink(protocol *model.Protocol) (string, error) {
	settings, err := m.GenerateVLESSConfig(protocol)
	if err != nil {
		return "", err
	}

	link := VLESSLink{
		ID:         settings.UUID,
		Flow:       settings.Flow,
		Encryption: "none",
		FP:         "chrome",
		Type:       "none",
		Host:       settings.Host,
		Path:       settings.Path,
		Port:       fmt.Sprintf("%d", protocol.Port),
	}

	if settings.TLS {
		link.Type = "tls"
	}

	return fmt.Sprintf("vless://%s@%s:%s?type=%s&host=%s&path=%s#%s",
		link.ID,
		link.Host,
		link.Port,
		link.Type,
		url.QueryEscape(link.Host),
		url.QueryEscape(link.Path),
		url.QueryEscape(protocol.Name),
	), nil
}

// GenerateTrojanLink 生成 Trojan 链接
func (m *ProtocolManager) GenerateTrojanLink(protocol *model.Protocol) (string, error) {
	settings, err := m.GenerateTrojanConfig(protocol)
	if err != nil {
		return "", err
	}

	link := TrojanLink{
		Password: settings.Password,
		Host:     settings.Host,
		Port:     fmt.Sprintf("%d", protocol.Port),
		Path:     settings.Path,
	}

	// 构建参数列表
	params := []string{"security=tls"}

	// 添加 SNI 参数，优先使用专门的SNI字段，如果不存在则使用Host
	sni := settings.SNI
	if sni == "" {
		sni = settings.Host
	}

	params = append(params, fmt.Sprintf("sni=%s", sni))

	// 添加 Path 参数，如果存在
	if settings.Path != "" {
		params = append(params, fmt.Sprintf("path=%s", url.QueryEscape(settings.Path)))
	}

	// 构建完整链接
	return fmt.Sprintf("trojan://%s@%s:%s?%s#%s",
		url.QueryEscape(link.Password),
		link.Host,
		link.Port,
		strings.Join(params, "&"),
		url.QueryEscape(protocol.Name),
	), nil
}

// GenerateShadowsocksLink 生成 Shadowsocks 链接
func (m *ProtocolManager) GenerateShadowsocksLink(protocol *model.Protocol) (string, error) {
	settings, err := m.GenerateShadowsocksConfig(protocol)
	if err != nil {
		return "", err
	}

	link := ShadowsocksLink{
		Method:   settings.Method,
		Password: settings.Password,
		Host:     settings.Host,
		Port:     fmt.Sprintf("%d", protocol.Port),
		Path:     settings.Path,
	}

	// 生成 Shadowsocks 链接
	ssLink := fmt.Sprintf("%s:%s@%s:%s",
		link.Method,
		link.Password,
		link.Host,
		link.Port,
	)

	// 添加插件参数
	if settings.AllowInsecure {
		ssLink += "?plugin=obfs-local;obfs=tls"
		if link.Path != "" {
			ssLink += ";obfs-host=" + url.QueryEscape(link.Path)
		}
	}

	return "ss://" + base64.URLEncoding.EncodeToString([]byte(ssLink)) + "#" + url.QueryEscape(protocol.Name), nil
}

// GenerateSubscriptionLink 生成订阅链接
func (m *ProtocolManager) GenerateSubscriptionLink(protocols []*model.Protocol) (string, error) {
	var links []string

	for _, protocol := range protocols {
		var link string
		var err error

		switch protocol.Type {
		case string(model.ProtocolVMess):
			link, err = m.GenerateVMessLink(protocol)
		case string(model.ProtocolVLESS):
			link, err = m.GenerateVLESSLink(protocol)
		case string(model.ProtocolTrojan):
			link, err = m.GenerateTrojanLink(protocol)
		case string(model.ProtocolShadowsocks):
			link, err = m.GenerateShadowsocksLink(protocol)
		default:
			continue
		}

		if err != nil {
			return "", err
		}

		links = append(links, link)
	}

	return base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n"))), nil
}
