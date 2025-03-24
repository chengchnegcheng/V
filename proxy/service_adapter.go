package proxy

import (
	"encoding/json"
	"fmt"
	"v/common"
	"v/model"
)

// ProxyServiceAdapter adapts the Service to implement model.ProxyService
type ProxyServiceAdapter struct {
	service *Service
}

// NewProxyServiceAdapter creates a new proxy service adapter
func NewProxyServiceAdapter(service *Service) *ProxyServiceAdapter {
	return &ProxyServiceAdapter{
		service: service,
	}
}

// CreateProxy implements model.ProxyService.CreateProxy
func (a *ProxyServiceAdapter) CreateProxy(proxy *model.Proxy) error {
	// Convert settings from string to map
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(proxy.Settings), &settings); err != nil {
		return fmt.Errorf("invalid settings: %v", err)
	}

	// Create a ProxyConfig
	proxyConfig := &common.ProxyConfig{
		ID:       proxy.ID,
		UserID:   proxy.UserID,
		Type:     proxy.Protocol,
		Port:     proxy.Port,
		Settings: proxy.Settings,
		Enabled:  proxy.Enabled,
	}

	return a.service.Create(proxyConfig)
}

// GetProxy implements model.ProxyService.GetProxy
func (a *ProxyServiceAdapter) GetProxy(id int64) (*model.Proxy, error) {
	// Get the proxy from the database
	p, err := a.service.GetProxyByID(id)
	if err != nil {
		return nil, err
	}

	// Convert common.Proxy to model.Proxy
	proxy := &model.Proxy{
		Base: model.Base{
			ID: p.ID,
		},
		UserID:   p.UserID,
		Protocol: p.Protocol,
		Port:     p.Port,
		Settings: p.Settings,
		Enabled:  p.Enabled,
	}

	return proxy, nil
}

// GetProxyByPort implements model.ProxyService.GetProxyByPort
func (a *ProxyServiceAdapter) GetProxyByPort(port int) (*model.Proxy, error) {
	// Get the proxy from the database
	p, err := a.service.GetProxyByPort(port)
	if err != nil {
		return nil, err
	}

	// Convert common.Proxy to model.Proxy
	proxy := &model.Proxy{
		Base: model.Base{
			ID: p.ID,
		},
		UserID:   p.UserID,
		Protocol: p.Protocol,
		Port:     p.Port,
		Settings: p.Settings,
		Enabled:  p.Enabled,
	}

	return proxy, nil
}

// ListProxies implements model.ProxyService.ListProxies
func (a *ProxyServiceAdapter) ListProxies(userID int64) ([]*model.Proxy, error) {
	// Get proxies from the database
	proxies, err := a.service.ListUserProxies(userID)
	if err != nil {
		return nil, err
	}

	// Convert common.Proxy to model.Proxy
	result := make([]*model.Proxy, len(proxies))
	for i, p := range proxies {
		result[i] = &model.Proxy{
			Base: model.Base{
				ID: p.ID,
			},
			UserID:   p.UserID,
			Protocol: p.Protocol,
			Port:     p.Port,
			Settings: p.Settings,
			Enabled:  p.Enabled,
		}
	}

	return result, nil
}

// UpdateProxy implements model.ProxyService.UpdateProxy
func (a *ProxyServiceAdapter) UpdateProxy(proxy *model.Proxy) error {
	// Convert settings from string to map
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(proxy.Settings), &settings); err != nil {
		return fmt.Errorf("invalid settings: %v", err)
	}

	// Create a ProxyConfig
	proxyConfig := &common.ProxyConfig{
		ID:       proxy.ID,
		UserID:   proxy.UserID,
		Type:     proxy.Protocol,
		Port:     proxy.Port,
		Settings: proxy.Settings,
		Enabled:  proxy.Enabled,
	}

	return a.service.Update(proxyConfig)
}

// DeleteProxy implements model.ProxyService.DeleteProxy
func (a *ProxyServiceAdapter) DeleteProxy(id int64) error {
	return a.service.Delete(id)
}

// EnableProxy implements model.ProxyService.EnableProxy
func (a *ProxyServiceAdapter) EnableProxy(id int64) error {
	return a.service.Enable(id)
}

// DisableProxy implements model.ProxyService.DisableProxy
func (a *ProxyServiceAdapter) DisableProxy(id int64) error {
	return a.service.Disable(id)
}

// GetProxyStats implements model.ProxyService.GetProxyStats
func (a *ProxyServiceAdapter) GetProxyStats(id int64) (*model.ProxyStats, error) {
	server, ok := a.service.servers[id]
	if !ok {
		return nil, fmt.Errorf("proxy not found: %d", id)
	}

	// Convert to model.ProxyStats format
	return &model.ProxyStats{
		ProxyID:   id,
		Upload:    server.GetUpload(),
		Download:  server.GetDownload(),
		Total:     server.GetUpload() + server.GetDownload(),
		UpdatedAt: server.GetLastActive(),
	}, nil
}
