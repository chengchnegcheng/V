package proxy

import (
	"encoding/json"
	"net/http"
	"strconv"
	"v/model"
	"v/utils"

	"github.com/gin-gonic/gin"
)

// CreateProxyRequest represents a request to create a proxy
type CreateProxyRequest struct {
	Protocol string                 `json:"protocol"`
	Port     int                    `json:"port"`
	Settings map[string]interface{} `json:"settings"`
	Enabled  bool                   `json:"enabled"`
}

// UpdateProxyRequest represents a request to update a proxy
type UpdateProxyRequest struct {
	Protocol string                 `json:"protocol"`
	Port     int                    `json:"port"`
	Settings map[string]interface{} `json:"settings"`
	Enabled  bool                   `json:"enabled"`
}

// Handler handles proxy-related requests
type Handler struct {
	service model.ProxyService
}

// NewHandler creates a new proxy handler
func NewHandler(service model.ProxyService) *Handler {
	return &Handler{service: service}
}

// CreateProxy creates a new proxy
func (h *Handler) CreateProxy(c *gin.Context) {
	var req CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c.Request.Context())
	settings, err := json.Marshal(req.Settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings"})
		return
	}

	proxy := &model.Proxy{
		UserID:   int64(userID),
		Protocol: req.Protocol,
		Port:     req.Port,
		Settings: string(settings),
		Enabled:  req.Enabled,
	}

	if err := h.service.CreateProxy(proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// GetProxy retrieves a proxy by ID
func (h *Handler) GetProxy(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.GetProxy(proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if int64(userID) != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// ListProxies lists all proxies for the current user
func (h *Handler) ListProxies(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxies, err := h.service.ListProxies(int64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxies)
}

// UpdateProxy updates a proxy
func (h *Handler) UpdateProxy(c *gin.Context) {
	var req UpdateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.GetProxy(proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if int64(userID) != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	settings, err := json.Marshal(req.Settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settings"})
		return
	}

	proxy.Protocol = req.Protocol
	proxy.Port = req.Port
	proxy.Settings = string(settings)
	proxy.Enabled = req.Enabled

	if err := h.service.UpdateProxy(proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// DeleteProxy deletes a proxy
func (h *Handler) DeleteProxy(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.GetProxy(proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if int64(userID) != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.DeleteProxy(proxyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy deleted successfully"})
}

// GetProxyStats retrieves a proxy's stats
func (h *Handler) GetProxyStats(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.GetProxy(proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if int64(userID) != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	stats, err := h.service.GetProxyStats(proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetProxiesByUser handles retrieving proxies for a user
func (h *Handler) GetProxiesByUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	proxies, err := h.service.ListProxies(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get proxies"})
		return
	}

	c.JSON(http.StatusOK, proxies)
}

// EnableProxy enables a proxy
func (h *Handler) EnableProxy(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.GetProxy(proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if int64(userID) != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.EnableProxy(proxyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy enabled successfully"})
}

// DisableProxy disables a proxy
func (h *Handler) DisableProxy(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.GetProxy(proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if int64(userID) != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.DisableProxy(proxyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy disabled successfully"})
}

// ListUserProxies lists all proxies for a user (admin only)
func (h *Handler) ListUserProxies(c *gin.Context) {
	// Get the admin status from the context, or check role
	// This is a placeholder - implement your actual admin check based on your auth system
	adminID := utils.GetUserIDFromContext(c.Request.Context())

	// For now, we'll assume any user can only access their own data
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// In a real system, you would check if adminID is actually an admin
	// For simplicity, we just check if the user is trying to access their own data
	if int64(adminID) != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied, you can only view your own proxies"})
		return
	}

	proxies, err := h.service.ListProxies(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxies)
}
