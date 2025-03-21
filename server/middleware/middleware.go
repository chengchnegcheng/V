package middleware

import (
	"time"
	"v/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		c.Set("latency", latency)
		c.Set("status_code", statusCode)

		// 记录访问日志
		c.MustGet("logger").(*logger.Logger).Info("HTTP Request",
			logger.Fields{
				"method":     c.Request.Method,
				"path":       path,
				"query":      query,
				"status":     statusCode,
				"latency":    latency,
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			})
	}
}

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}

		// TODO: 验证 token
		// user, err := auth.ValidateToken(token)
		// if err != nil {
		// 	c.JSON(401, gin.H{
		// 		"error": "Invalid token",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// c.Set("user", user)
		c.Next()
	}
}
