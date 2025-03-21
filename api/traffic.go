package api

import (
	"net/http"
	"strconv"
	"time"

	"v/logger"
	"v/traffic"

	"github.com/gin-gonic/gin"
)

// TrafficHandler 流量统计API处理器
type TrafficHandler struct {
	log *logger.Logger
	mgr *traffic.Manager
}

// NewTrafficHandler 创建流量统计处理器
func NewTrafficHandler(log *logger.Logger, mgr *traffic.Manager) *TrafficHandler {
	return &TrafficHandler{
		log: log,
		mgr: mgr,
	}
}

// RegisterRoutes 注册路由
func (h *TrafficHandler) RegisterRoutes(router *gin.RouterGroup) {
	trafficGroup := router.Group("/traffic")
	{
		trafficGroup.GET("/stats", h.GetTrafficStats)
		trafficGroup.GET("/user/:id", h.GetUserTraffic)
		trafficGroup.GET("/daily", h.GetDailyTraffic)
		trafficGroup.GET("/limits", h.GetTrafficLimits)
		trafficGroup.POST("/limits/user/:id", h.UpdateUserTrafficLimit)
	}
}

// GetTrafficStats 获取总流量统计
func (h *TrafficHandler) GetTrafficStats(c *gin.Context) {
	stats := h.mgr.GetTrafficStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUserTraffic 获取指定用户的流量统计
func (h *TrafficHandler) GetUserTraffic(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的用户ID",
			"error":   err.Error(),
		})
		return
	}

	traffic, err := h.mgr.GetUserTraffic(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户流量失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    traffic,
	})
}

// GetDailyTraffic 获取每日流量统计
func (h *TrafficHandler) GetDailyTraffic(c *gin.Context) {
	// 解析查询参数
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	userIDStr := c.Query("user_id")

	var startDate, endDate time.Time
	var err error
	var userID int64 = 0

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的开始日期",
				"error":   err.Error(),
			})
			return
		}
	} else {
		// 默认为7天前
		startDate = time.Now().AddDate(0, 0, -7)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的结束日期",
				"error":   err.Error(),
			})
			return
		}
	} else {
		// 默认为今天
		endDate = time.Now()
	}

	if userIDStr != "" {
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的用户ID",
				"error":   err.Error(),
			})
			return
		}
	}

	// 获取每日流量
	dailyTraffic, err := h.mgr.GetDailyTraffic(startDate, endDate, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取每日流量失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dailyTraffic,
	})
}

// GetTrafficLimits 获取流量限制
func (h *TrafficHandler) GetTrafficLimits(c *gin.Context) {
	limits, err := h.mgr.GetTrafficLimits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取流量限制失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    limits,
	})
}

// UpdateUserTrafficLimit 更新用户流量限制
func (h *TrafficHandler) UpdateUserTrafficLimit(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的用户ID",
			"error":   err.Error(),
		})
		return
	}

	var request struct {
		Limit int64 `json:"limit" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数",
			"error":   err.Error(),
		})
		return
	}

	if err := h.mgr.UpdateUserTrafficLimit(userID, request.Limit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新用户流量限制失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户流量限制已更新",
	})
}
