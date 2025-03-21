package proxy

import (
	"net/http"
	"v/model"

	"github.com/gin-gonic/gin"
)

// Middleware handles proxy-related middleware
type Middleware struct {
	proxyService *Service
}

// NewMiddleware creates a new proxy middleware
func NewMiddleware(proxyService *Service) *Middleware {
	return &Middleware{proxyService: proxyService}
}

// RequireProxyOwner ensures that the request is from the proxy owner
func (m *Middleware) RequireProxyOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get proxy ID from URL parameter
		id, err := c.GetInt64("id")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
			c.Abort()
			return
		}

		// Get proxy
		proxy, err := m.proxyService.GetProxy(id)
		if err != nil {
			switch err {
			case ErrProxyNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "proxy not found"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get proxy"})
			}
			c.Abort()
			return
		}

		// Get user from context
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Check if user owns the proxy
		if user.(*model.User).ID != proxy.UserID && !user.(*model.User).IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// Set proxy in context
		c.Set("proxy", proxy)
		c.Next()
	}
}

// RequireProxyEnabled ensures that the proxy is enabled
func (m *Middleware) RequireProxyEnabled() gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy, exists := c.Get("proxy")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "proxy not found in context"})
			c.Abort()
			return
		}

		if !proxy.(*model.Proxy).Enabled {
			c.JSON(http.StatusForbidden, gin.H{"error": "proxy is disabled"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireProxyPermission ensures that the user has permission to manage proxies
func (m *Middleware) RequireProxyPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Admin users have all permissions
		if user.(*model.User).IsAdmin {
			c.Next()
			return
		}

		// TODO: Implement permission checking logic
		c.Next()
	}
}
