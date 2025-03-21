package api

import (
	"net/http"
	"strconv"
	"time"

	"v/logger"
	"v/model"

	"github.com/gin-gonic/gin"
)

// LogHandler 日志处理器
type LogHandler struct {
	logger   *logger.Logger
	db       model.DB
	manager  *logger.Manager
	analyzer *logger.LogAnalyzer
}

// NewLogHandler 创建日志处理器
func NewLogHandler(log *logger.Logger, db model.DB, manager *logger.Manager, analyzer *logger.LogAnalyzer) *LogHandler {
	return &LogHandler{
		logger:   log,
		db:       db,
		manager:  manager,
		analyzer: analyzer,
	}
}

// RegisterRoutes 注册路由
func (h *LogHandler) RegisterRoutes(router *gin.RouterGroup) {
	logRouter := router.Group("/logs")
	{
		logRouter.GET("", h.ListLogs)
		logRouter.GET("/search", h.SearchLogs)
		logRouter.GET("/statistics", h.GetLogStatistics)
		logRouter.GET("/errors", h.GetErrorLogs)
		logRouter.POST("/export", h.ExportLogs)
		logRouter.DELETE("", h.CleanupLogs)
	}
}

// ListLogs 获取日志列表
func (h *LogHandler) ListLogs(c *gin.Context) {
	// 解析查询参数
	query := &model.LogQuery{
		Page:     1,
		PageSize: 20,
	}

	if level := c.Query("level"); level != "" {
		query.Level = level
	}
	if module := c.Query("module"); module != "" {
		query.Module = module
	}
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = t
		}
	}
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseInt(userID, 10, 64); err == nil {
			query.UserID = id
		}
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			query.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			query.PageSize = ps
		}
	}

	// 查询日志
	logs, err := h.manager.ListLogs(query)
	if err != nil {
		h.logger.Error("Failed to list logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取日志失败",
			"error":   err.Error(),
		})
		return
	}

	// 获取总数
	total, err := h.manager.GetTotalLogs(query)
	if err != nil {
		h.logger.Error("Failed to get total logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取日志总数失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":       logs,
			"total":      total,
			"page":       query.Page,
			"page_size":  query.PageSize,
			"total_page": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
		},
	})
}

// SearchLogs 搜索日志文件
func (h *LogHandler) SearchLogs(c *gin.Context) {
	// 解析查询参数
	query := &logger.LogQuery{}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = t
		}
	}
	if level := c.Query("level"); level != "" {
		query.Level = level
	}
	if message := c.Query("message"); message != "" {
		query.Message = message
	}
	if file := c.Query("file"); file != "" {
		query.File = file
	}
	if function := c.Query("function"); function != "" {
		query.Function = function
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			query.Limit = l
		}
	} else {
		query.Limit = 100 // 默认限制100条
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			query.Offset = o
		}
	}

	// 搜索日志
	logs, err := h.analyzer.SearchLogs(query)
	if err != nil {
		h.logger.Error("Failed to search logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "搜索日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":  logs,
			"total": len(logs),
		},
	})
}

// GetLogStatistics 获取日志统计信息
func (h *LogHandler) GetLogStatistics(c *gin.Context) {
	// 解析查询参数
	var startTime, endTime time.Time

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	} else {
		// 默认为过去7天
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	} else {
		endTime = time.Now()
	}

	// 获取统计信息
	stats, err := h.analyzer.GetLogStats(startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get log statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取日志统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetErrorLogs 获取错误日志
func (h *LogHandler) GetErrorLogs(c *gin.Context) {
	// 解析参数
	days := 7 // 默认7天
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	limit := 100 // 默认100条
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 获取错误日志
	logs, err := h.analyzer.GetErrorLogs(days, limit)
	if err != nil {
		h.logger.Error("Failed to get error logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取错误日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":  logs,
			"total": len(logs),
		},
	})
}

// ExportLogs 导出日志
func (h *LogHandler) ExportLogs(c *gin.Context) {
	// 解析请求
	var req struct {
		Level     string    `json:"level"`
		Module    string    `json:"module"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		UserID    int64     `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 构造查询参数
	query := &model.LogQuery{
		Level:     req.Level,
		Module:    req.Module,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		UserID:    req.UserID,
	}

	// 导出日志
	filePath, err := h.manager.ExportLogs(query)
	if err != nil {
		h.logger.Error("Failed to export logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "导出日志失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"file_path": filePath,
		},
	})
}

// CleanupLogs 清理日志
func (h *LogHandler) CleanupLogs(c *gin.Context) {
	// 解析参数
	days := 30 // 默认清理30天前的日志
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// 清理数据库日志
	if err := h.manager.DeleteLogs(&model.LogQuery{
		EndTime: time.Now().AddDate(0, 0, -days),
	}); err != nil {
		h.logger.Error("Failed to cleanup logs in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清理数据库日志失败",
			"error":   err.Error(),
		})
		return
	}

	// 清理日志文件
	if err := h.analyzer.TruncateLogs(time.Now().AddDate(0, 0, -days)); err != nil {
		h.logger.Error("Failed to truncate log files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清理日志文件失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "日志清理成功",
	})
}
