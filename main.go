package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.New()
	r.Use(gin.Recovery())

	// 添加中间件
	r.Use(func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		// 日志格式
		fmt.Printf("[GIN] %v | %3d | %13v | %15s | %s | %s\n",
			endTime.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)
	})

	// 路由组
	apiGroup := r.Group("/api")
	{
		// 健康检查
		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})

		// 系统信息
		apiGroup.GET("/system/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":   "running",
				"version":  "1.0.0",
				"uptime":   time.Now().Format(time.RFC3339),
				"hostname": "V-Server",
				"system": gin.H{
					"cpu":    0,
					"memory": 0,
					"disk":   0,
				},
			})
		})

		// 用户列表
		apiGroup.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"users": []gin.H{
					{
						"id":       1,
						"username": "admin",
						"email":    "admin@example.com",
						"role":     "admin",
						"created":  time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					},
				},
			})
		})

		// 流量统计
		apiGroup.GET("/traffic", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"total_upload":   0,
				"total_download": 0,
				"active_users":   1,
				"last_updated":   time.Now().Format(time.RFC3339),
			})
		})

		// 协议列表
		apiGroup.GET("/protocols", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"protocols": []gin.H{
					{
						"id":         1,
						"name":       "默认VMess协议",
						"type":       "vmess",
						"port":       10086,
						"enabled":    true,
						"user_id":    1,
						"created_at": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					},
					{
						"id":         2,
						"name":       "默认Trojan协议",
						"type":       "trojan",
						"port":       443,
						"enabled":    true,
						"user_id":    1,
						"created_at": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
					},
				},
			})
		})
	}

	// 添加默认首页
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "V 多协议代理面板 - 已准备就绪")
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 启动HTTP服务器
	go func() {
		fmt.Println("V 服务已启动，监听端口: 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器启动失败: %v\n", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	fmt.Println("正在关闭服务...")

	// 创建带超时的上下文来关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("关闭服务器错误: %v\n", err)
	}

	fmt.Println("服务已安全关闭")
}
