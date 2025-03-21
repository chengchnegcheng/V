package api

import (
	"net/http"

	"v/logger"
	stg "v/settings"

	"github.com/gin-gonic/gin"
)

// SettingsHandler 系统设置API处理器
type SettingsHandler struct {
	log      *logger.Logger
	settings *stg.Manager
}

// NewSettingsHandler 创建系统设置处理器
func NewSettingsHandler(log *logger.Logger, settings *stg.Manager) *SettingsHandler {
	return &SettingsHandler{
		log:      log,
		settings: settings,
	}
}

// RegisterRoutes 注册路由
func (h *SettingsHandler) RegisterRoutes(router *gin.RouterGroup) {
	settingsGroup := router.Group("/settings")
	{
		settingsGroup.GET("", h.GetSettings)
		settingsGroup.PUT("", h.UpdateSettings)
		settingsGroup.GET("/sections/:section", h.GetSectionSettings)
		settingsGroup.PUT("/sections/:section", h.UpdateSectionSettings)
		settingsGroup.POST("/backup", h.BackupSettings)
		settingsGroup.POST("/restore", h.RestoreSettings)
	}
}

// GetSettings 获取所有设置
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings := h.settings.Get()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateSettings 更新所有设置
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var settings stg.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	if err := h.settings.Update(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新设置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置已更新",
	})
}

// GetSectionSettings 获取指定部分的设置
func (h *SettingsHandler) GetSectionSettings(c *gin.Context) {
	section := c.Param("section")
	settings := h.settings.Get()

	var sectionData interface{}
	switch section {
	case "site":
		sectionData = settings.Site
	case "admin":
		sectionData = settings.Admin
	case "ssl":
		sectionData = settings.SSL
	case "notification":
		sectionData = settings.Notification
	case "monitor":
		sectionData = settings.Monitor
	case "traffic":
		sectionData = settings.Traffic
	case "log":
		sectionData = settings.Log
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的设置部分",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    sectionData,
	})
}

// UpdateSectionSettings 更新指定部分的设置
func (h *SettingsHandler) UpdateSectionSettings(c *gin.Context) {
	section := c.Param("section")
	settings := h.settings.Get()

	switch section {
	case "site":
		var siteSettings stg.SiteSettings
		if err := c.ShouldBindJSON(&siteSettings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求参数",
				"error":   err.Error(),
			})
			return
		}
		settings.Site = siteSettings
	case "admin":
		var adminSettings stg.AdminSettings
		if err := c.ShouldBindJSON(&adminSettings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求参数",
				"error":   err.Error(),
			})
			return
		}
		settings.Admin = adminSettings
	case "ssl":
		var sslSettings stg.SSLSettings
		if err := c.ShouldBindJSON(&sslSettings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求参数",
				"error":   err.Error(),
			})
			return
		}
		settings.SSL = sslSettings
	case "notification":
		var notificationSettings stg.NotificationSettings
		if err := c.ShouldBindJSON(&notificationSettings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求参数",
				"error":   err.Error(),
			})
			return
		}
		settings.Notification = notificationSettings
	case "monitor":
		var monitorSettings stg.MonitorSettings
		if err := c.ShouldBindJSON(&monitorSettings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求参数",
				"error":   err.Error(),
			})
			return
		}
		settings.Monitor = monitorSettings
	case "traffic":
		var trafficSettings stg.TrafficSettings
		if err := c.ShouldBindJSON(&trafficSettings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求参数",
				"error":   err.Error(),
			})
			return
		}
		settings.Traffic = trafficSettings
	case "log":
		var logSettings stg.LogSettings
		if err := c.ShouldBindJSON(&logSettings); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的请求参数",
				"error":   err.Error(),
			})
			return
		}
		settings.Log = logSettings
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的设置部分",
		})
		return
	}

	if err := h.settings.Update(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新设置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置已更新",
	})
}

// BackupSettings 备份设置
func (h *SettingsHandler) BackupSettings(c *gin.Context) {
	backupPath, err := h.settings.Backup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "备份设置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置已备份",
		"data": gin.H{
			"backup_path": backupPath,
		},
	})
}

// RestoreSettings 恢复设置
func (h *SettingsHandler) RestoreSettings(c *gin.Context) {
	var req struct {
		BackupPath string `json:"backup_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	if err := h.settings.Restore(req.BackupPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "恢复设置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设置已恢复",
	})
}
