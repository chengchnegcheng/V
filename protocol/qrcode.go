package protocol

import (
	"encoding/base64"
	"errors"

	"v/model"
)

// ErrUnsupportedProtocol 不支持的协议类型错误
var ErrUnsupportedProtocol = errors.New("unsupported protocol type")

// GenerateQRCode 生成协议配置的二维码
func (m *ProtocolManager) GenerateQRCode(protocol *model.Protocol) (string, error) {
	// 生成协议链接
	var link string
	var err error

	switch protocol.Type {
	case "vmess":
		link, err = m.GenerateVMessLink(protocol)
	case "vless":
		link, err = m.GenerateVLESSLink(protocol)
	case "trojan":
		link, err = m.GenerateTrojanLink(protocol)
	case "shadowsocks":
		link, err = m.GenerateShadowsocksLink(protocol)
	default:
		return "", ErrUnsupportedProtocol
	}

	if err != nil {
		return "", err
	}

	// 占位实现，实际开发时使用正确的QR代码库
	// 这里简单返回协议链接的base64编码，表示QR内容
	return "data:text/plain;base64," + base64.StdEncoding.EncodeToString([]byte(link)), nil
}

// GenerateSubscriptionQRCode 生成订阅链接的二维码
func (m *ProtocolManager) GenerateSubscriptionQRCode(protocols []*model.Protocol) (string, error) {
	// 生成订阅链接
	link, err := m.GenerateSubscriptionLink(protocols)
	if err != nil {
		return "", err
	}

	// 占位实现，实际开发时使用正确的QR代码库
	// 这里简单返回协议链接的base64编码，表示QR内容
	return "data:text/plain;base64," + base64.StdEncoding.EncodeToString([]byte(link)), nil
}
