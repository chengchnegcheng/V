package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
)

// HandleHealth 处理健康检查请求
func HandleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
