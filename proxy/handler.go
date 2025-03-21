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
		UserID:   userID,
		Protocol: req.Protocol,
		Port:     req.Port,
		Settings: string(settings),
		Enabled:  req.Enabled,
	}

	if err := h.service.Create(proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// GetProxy retrieves a proxy by ID
func (h *Handler) GetProxy(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.Get(uint(proxyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userID != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// ListProxies lists all proxies for the current user
func (h *Handler) ListProxies(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxies, err := h.service.GetByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxies)
}

// UpdateProxy updates a proxy
func (h *Handler) UpdateProxy(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.Get(uint(proxyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userID != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req UpdateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if err := h.service.Update(proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// DeleteProxy deletes a proxy
func (h *Handler) DeleteProxy(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.Get(uint(proxyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userID != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.Delete(uint(proxyID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "proxy deleted"})
}

// GetProxyStats retrieves traffic statistics for a proxy
func (h *Handler) GetProxyStats(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c.Request.Context())
	proxyID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	proxy, err := h.service.Get(uint(proxyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userID != proxy.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	upload, download := proxy.Upload, proxy.Download
	c.JSON(http.StatusOK, gin.H{
		"upload":   upload,
		"download": download,
	})
}

// GetProxiesByUser handles retrieving proxies for a user
func (h *Handler) GetProxiesByUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	proxies, err := h.service.GetByUser(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get proxies"})
		return
	}

	c.JSON(http.StatusOK, proxies)
}

// EnableProxy handles proxy enabling
func (h *Handler) EnableProxy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	if err := h.service.Enable(uint(id)); err != nil {
		switch err {
		case ErrProxyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "proxy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable proxy"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "proxy enabled successfully"})
}

// DisableProxy handles proxy disabling
func (h *Handler) DisableProxy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid proxy ID"})
		return
	}

	if err := h.service.Disable(uint(id)); err != nil {
		switch err {
		case ErrProxyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "proxy not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable proxy"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "proxy disabled successfully"})
}
