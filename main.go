package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

	// API路由组
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

	// 检查是否存在dist目录
	distDir := "./web/dist"
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		// 如果不存在dist目录，使用一个临时的HTML页面
		r.GET("/", func(c *gin.Context) {
			c.Header("Content-Type", "text/html")
			c.String(http.StatusOK, `
			<!DOCTYPE html>
			<html>
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>V 多协议代理面板</title>
				<style>
					body {
						font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
						display: flex;
						flex-direction: column;
						align-items: center;
						justify-content: center;
						height: 100vh;
						margin: 0;
						background-color: #f5f7fa;
						color: #333;
					}
					.container {
						text-align: center;
						padding: 2rem;
						background-color: white;
						border-radius: 10px;
						box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
						max-width: 800px;
						width: 90%;
					}
					h1 {
						color: #409eff;
						margin-bottom: 1rem;
					}
					p {
						margin-bottom: 2rem;
						line-height: 1.6;
					}
					.btn {
						background-color: #409eff;
						color: white;
						border: none;
						padding: 10px 20px;
						border-radius: 4px;
						cursor: pointer;
						font-size: 16px;
						text-decoration: none;
						display: inline-block;
						margin: 0 10px;
					}
					.btn:hover {
						background-color: #66b1ff;
					}
					.status {
						margin-top: 2rem;
						padding: 15px;
						background-color: #f0f9eb;
						border-radius: 4px;
						color: #67c23a;
						font-weight: bold;
					}
					.error {
						margin-top: 1rem;
						color: #f56c6c;
					}
					.info {
						font-size: 0.9rem;
						color: #909399;
						margin-top: 2rem;
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>V 多协议代理面板</h1>
					<p>V 是一个功能强大的多协议代理面板，支持 vmess、vless、trojan、shadowsocks 等多种协议。</p>
					<p>前端资源尚未编译，请按照以下步骤完成设置：</p>
					<ol style="text-align: left;">
						<li>确保已安装 Node.js 和 npm</li>
						<li>进入 web 目录: <code>cd web</code></li>
						<li>安装依赖: <code>npm install</code></li>
						<li>构建前端: <code>npm run build</code></li>
						<li>重启服务</li>
					</ol>
					<div class="status">服务器运行正常，API 接口可用</div>
					<p class="info">当前版本: 1.0.0 | 服务器时间: `+time.Now().Format("2006-01-02 15:04:05")+`</p>
				</div>
			</body>
			</html>
			`)
		})
	} else {
		// 如果存在dist目录，则提供静态文件服务
		r.StaticFS("/assets", http.Dir(filepath.Join(distDir, "assets")))
		r.StaticFile("/favicon.ico", filepath.Join(distDir, "favicon.ico"))

		// 处理所有前端路由
		r.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(distDir, "index.html"))
		})
	}

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
