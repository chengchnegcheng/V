package api

import (
	"net/http"

	"v/logger"
	"v/model"
	"v/monitor"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统信息API处理器
type SystemHandler struct {
	log     *logger.Logger
	monitor *monitor.MonitorManager
}

// NewSystemHandler 创建系统信息处理器
func NewSystemHandler(log *logger.Logger, monitor *monitor.MonitorManager) *SystemHandler {
	return &SystemHandler{
		log:     log,
		monitor: monitor,
	}
}

// RegisterRoutes 注册路由
func (h *SystemHandler) RegisterRoutes(router *gin.RouterGroup) {
	systemGroup := router.Group("/system")
	{
		systemGroup.GET("/info", h.GetSystemInfo)
		systemGroup.GET("/stats", h.GetSystemStats)
		systemGroup.GET("/alerts", h.GetAlerts)
		systemGroup.POST("/alerts/test", h.SendTestAlert)
	}
}

// GetSystemInfo 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	info := model.GetSystemInfo()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// GetSystemStats 获取系统统计信息
func (h *SystemHandler) GetSystemStats(c *gin.Context) {
	stats, err := model.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取系统统计信息失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetAlerts 获取告警记录
func (h *SystemHandler) GetAlerts(c *gin.Context) {
	// 这里需要实现获取告警记录的逻辑
	// 暂时返回空数组
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []model.AlertRecord{},
	})
}

// SendTestAlert 发送测试告警
func (h *SystemHandler) SendTestAlert(c *gin.Context) {
	err := h.monitor.SendTestAlert()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "发送测试告警失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试告警已发送",
	})
}
