package handlers

import (
	"net/http"

	"v/logger"
	"v/monitor"

	"github.com/gin-gonic/gin"
)

var monitorMgr *monitor.Manager

// InitMonitorHandlers 初始化系统监控处理器
func InitMonitorHandlers(log *logger.Logger) {
	monitorMgr = monitor.New(log)
	if err := monitorMgr.Start(); err != nil {
		log.Error("Failed to start monitor", logger.Fields{
			"error": err.Error(),
		})
	}
}

// HandleGetSystemStats 处理获取系统状态的请求
func HandleGetSystemStats(c *gin.Context) {
	// 获取系统状态
	stats, err := monitorMgr.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get system stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
