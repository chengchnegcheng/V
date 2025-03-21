package handlers

import (
	"net/http"
	"strconv"
	"time"

	"v/logger"
	"v/model"

	"github.com/gin-gonic/gin"
)

var logMgr *logger.Manager

// InitLogHandlers 初始化日志处理器
func InitLogHandlers(log *logger.Logger, db model.DB) {
	logMgr = logger.NewManager(log, db)
}

// HandleListLogs 处理获取日志列表的请求
func HandleListLogs(c *gin.Context) {
	query := &model.LogQuery{
		Page:     1,
		PageSize: 20,
	}

	// 解析查询参数
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

	// 获取日志列表
	logs, err := logMgr.ListLogs(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list logs",
		})
		return
	}

	// 获取总数
	total, err := logMgr.GetTotalLogs(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get total logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
	})
}

// HandleExportLogs 处理导出日志的请求
func HandleExportLogs(c *gin.Context) {
	query := &model.LogQuery{
		Page:     1,
		PageSize: 1000, // 导出时使用较大的页面大小
	}

	// 解析查询参数
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

	// 导出日志
	filepath, err := logMgr.ExportLogs(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to export logs",
		})
		return
	}

	// 发送文件
	c.File(filepath)
}

// HandleDeleteLogs 处理删除日志的请求
func HandleDeleteLogs(c *gin.Context) {
	query := &model.LogQuery{}

	// 解析查询参数
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

	// 删除日志
	if err := logMgr.DeleteLogs(query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logs deleted successfully",
	})
}
