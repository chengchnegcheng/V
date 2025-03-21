package api

import (
	"net/http"
	"strconv"

	"v/logger"
	"v/model"
	"v/protocol"

	"github.com/gin-gonic/gin"
)

// ProtocolHandler 协议管理API处理器
type ProtocolHandler struct {
	log *logger.Logger
	mgr *protocol.Manager
}

// NewProtocolHandler 创建协议管理处理器
func NewProtocolHandler(log *logger.Logger, mgr *protocol.Manager) *ProtocolHandler {
	return &ProtocolHandler{
		log: log,
		mgr: mgr,
	}
}

// RegisterRoutes 注册路由
func (h *ProtocolHandler) RegisterRoutes(router *gin.RouterGroup) {
	protocolGroup := router.Group("/protocols")
	{
		protocolGroup.GET("", h.ListProtocols)
		protocolGroup.GET("/:id", h.GetProtocol)
		protocolGroup.POST("", h.CreateProtocol)
		protocolGroup.PUT("/:id", h.UpdateProtocol)
		protocolGroup.DELETE("/:id", h.DeleteProtocol)
		protocolGroup.GET("/stats", h.GetProtocolStats)
		protocolGroup.GET("/types", h.GetProtocolTypes)
	}
}

// ListProtocols 列出所有协议
func (h *ProtocolHandler) ListProtocols(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	protocols, err := h.mgr.ListProtocols(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取协议列表失败",
			"error":   err.Error(),
		})
		return
	}

	totalCount, err := h.mgr.GetTotalProtocols()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取协议总数失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"protocols":   protocols,
			"total":       totalCount,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (totalCount + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetProtocol 获取指定协议
func (h *ProtocolHandler) GetProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的协议ID",
			"error":   err.Error(),
		})
		return
	}

	protocol, err := h.mgr.GetProtocol(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取协议信息失败",
			"error":   err.Error(),
		})
		return
	}

	if protocol == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "协议不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    protocol,
	})
}

// CreateProtocol 创建协议
func (h *ProtocolHandler) CreateProtocol(c *gin.Context) {
	var protocol model.Protocol
	if err := c.ShouldBindJSON(&protocol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	if err := h.mgr.CreateProtocol(&protocol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建协议失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "协议创建成功",
		"data":    protocol,
	})
}

// UpdateProtocol 更新协议
func (h *ProtocolHandler) UpdateProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的协议ID",
			"error":   err.Error(),
		})
		return
	}

	var protocol model.Protocol
	if err := c.ShouldBindJSON(&protocol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	protocol.ID = id
	if err := h.mgr.UpdateProtocol(&protocol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新协议失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "协议更新成功",
		"data":    protocol,
	})
}

// DeleteProtocol 删除协议
func (h *ProtocolHandler) DeleteProtocol(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的协议ID",
			"error":   err.Error(),
		})
		return
	}

	if err := h.mgr.DeleteProtocol(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除协议失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "协议删除成功",
	})
}

// GetProtocolStats 获取协议统计
func (h *ProtocolHandler) GetProtocolStats(c *gin.Context) {
	stats, err := h.mgr.GetProtocolStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取协议统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetProtocolTypes 获取协议类型
func (h *ProtocolHandler) GetProtocolTypes(c *gin.Context) {
	types := h.mgr.GetSupportedProtocolTypes()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    types,
	})
}
