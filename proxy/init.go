package proxy

import "v/model"

// DefaultServiceAdapter is the default proxy service adapter
var DefaultServiceAdapter model.ProxyService

// Initialize initializes the proxy service
func Initialize() {
	// Create the adapter for the default service
	DefaultServiceAdapter = NewProxyServiceAdapter(DefaultService)
}
