package handlers

import (
	"net/http"

	"v/logger"
	"v/monitor"

	"github.com/gin-gonic/gin"
)

var monitorInstance *monitor.Monitor

// InitMonitorHandlers 初始化系统监控处理器
func InitMonitorHandlers(log *logger.Logger) {
	monitorInstance = monitor.NewMonitor(log)
	if err := monitorInstance.Start(); err != nil {
		log.Error("Failed to start monitor", logger.Fields{
			"error": err.Error(),
		})
	}
}

// HandleGetSystemStats 处理获取系统状态的请求
func HandleGetSystemStats(c *gin.Context) {
	// 获取系统状态
	if monitorInstance == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Monitor not initialized",
		})
		return
	}

	stats := monitorInstance.GetStats()
	c.JSON(http.StatusOK, stats)
}
