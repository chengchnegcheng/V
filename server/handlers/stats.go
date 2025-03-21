package handlers

import (
	"net/http"
	"strconv"
	"time"

	"v/logger"
	"v/model"
	"v/stats"

	"github.com/gin-gonic/gin"
)

var statsMgr *stats.Manager

// InitStatsHandlers 初始化流量统计处理器
func InitStatsHandlers(log *logger.Logger, db model.DB) {
	statsMgr = stats.NewStatsManager(log, db)
	if err := statsMgr.Start(); err != nil {
		log.Error("Failed to start stats manager", logger.Fields{
			"error": err.Error(),
		})
	}
}

// HandleGetUserStats 获取用户流量统计
func HandleGetUserStats(c *gin.Context) {
	// 获取用户ID
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// 获取日期范围
	startDate, err := time.Parse(time.RFC3339, c.Query("start_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start date",
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end date",
		})
		return
	}

	// 获取统计数据
	userStats, err := statsMgr.GetUserStats(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": userStats,
	})
}

// HandleGetProtocolStats 获取协议流量统计
func HandleGetProtocolStats(c *gin.Context) {
	// 获取协议ID
	protocolID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid protocol ID",
		})
		return
	}

	// 获取日期范围
	startDate, err := time.Parse(time.RFC3339, c.Query("start_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start date",
		})
		return
	}

	endDate, err := time.Parse(time.RFC3339, c.Query("end_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end date",
		})
		return
	}

	// 获取统计数据
	protocolStats, err := statsMgr.GetProtocolStats(protocolID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": protocolStats,
	})
}

// HandleUpdateProtocolTraffic 更新协议流量
func HandleUpdateProtocolTraffic(c *gin.Context) {
	// 获取协议ID
	protocolID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid protocol ID",
		})
		return
	}

	// 获取流量数据
	var req struct {
		Upload   int64 `json:"upload" binding:"required"`
		Download int64 `json:"download" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// 更新流量
	if err := statsMgr.UpdateProtocolTraffic(protocolID, req.Upload, req.Download); err != nil {
		if err == model.ErrTrafficLimitExceeded {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Traffic limit exceeded",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Traffic updated successfully",
	})
}
