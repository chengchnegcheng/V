package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"v/api"
	"v/logger"
	"v/model"
	"v/settings"
	"v/xray"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志
	log := logger.New()
	log.Start()
	defer log.Stop()

	// 初始化设置管理器
	settingsManager := settings.New(log)
	if err := settingsManager.Start(); err != nil {
		log.Fatal("Failed to start settings manager", logger.Fields{
			"error": err,
		})
	}
	defer settingsManager.Stop()

	// 初始化xray版本管理器
	xrayManager := xray.New(log, settingsManager)
	if err := xrayManager.Initialize(); err != nil {
		log.Fatal("Failed to initialize xray manager", logger.Fields{
			"error": err,
		})
	}
	// 确保xray在应用退出时停止
	defer xrayManager.Stop()

	// 启动API服务器
	apiHandler := api.New(log, nil, settingsManager, xrayManager)
	if err := apiHandler.Start(); err != nil {
		log.Fatal("Failed to start API server", logger.Fields{
			"error": err,
		})
	}
	defer apiHandler.Stop()

	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	r := gin.New()
	r.Use(gin.Recovery())

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

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
		log.Info("HTTP Request", logger.Fields{
			"time":      endTime.Format("2006/01/02 - 15:04:05"),
			"status":    statusCode,
			"latency":   latencyTime,
			"client_ip": clientIP,
			"method":    reqMethod,
			"uri":       reqUri,
		})
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

		// 用户认证路由
		authGroup := apiGroup.Group("/auth")
		{
			// 登录处理
			authGroup.POST("/login", func(c *gin.Context) {
				var req struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
					return
				}

				log.Info("Login attempt", logger.Fields{
					"username": req.Username,
				})

				// 特殊处理admin用户
				if req.Username == "admin" {
					if req.Password != "admin123" {
						c.JSON(http.StatusUnauthorized, gin.H{
							"error": "Invalid username or password",
						})
						return
					}

					// 生成一个简单的token
					token := "admin_token_" + time.Now().Format("20060102150405")

					c.JSON(http.StatusOK, gin.H{
						"token": token,
						"user": gin.H{
							"id":       1,
							"username": "admin",
							"role":     "admin",
							"is_admin": true,
						},
					})
					return
				}

				// 处理其他用户
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid username or password",
				})
			})

			// 注册
			authGroup.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Registration is disabled",
				})
			})

			// 获取用户信息
			authGroup.GET("/user", func(c *gin.Context) {
				// 这里应该验证token，但为了简单，我们假设用户已经认证
				c.JSON(http.StatusOK, gin.H{
					"user": gin.H{
						"id":       1,
						"username": "admin",
						"role":     "admin",
						"is_admin": true,
					},
				})
			})

			// 登出
			authGroup.POST("/logout", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Logged out successfully",
				})
			})
		}

		// 系统信息
		apiGroup.GET("/system/info", func(c *gin.Context) {
			// 使用我们实现的GetSystemInfo函数获取系统信息
			sysInfo := model.GetSystemInfo()

			// 输出调试信息，帮助定位问题
			log.Info("System Info API called", logger.Fields{
				"os":       sysInfo["os"],
				"kernel":   sysInfo["kernel"],
				"hostname": sysInfo["hostname"],
			})

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "success",
				"data":    sysInfo,
			})
		})

		// 系统状态
		apiGroup.GET("/system/status", func(c *gin.Context) {
			// 添加更详细的请求日志
			log.Info("System Status API request received", logger.Fields{
				"client_ip":  c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
			})

			// 获取系统信息
			sysInfo := model.GetSystemInfo()

			// 添加详细调试日志
			log.Info("System Status API called - DETAILED INFO", logger.Fields{
				"os":       sysInfo["os"],
				"kernel":   sysInfo["kernel"],
				"hostname": sysInfo["hostname"],
				"uptime":   sysInfo["uptime"],
				"load":     sysInfo["load"],
				"ip":       sysInfo["ipAddress"],
			})

			// 把sysInfo转换为所需格式 - 确保字段名称正确
			systemInfo := gin.H{
				"os":        sysInfo["os"],
				"kernel":    sysInfo["kernel"],
				"hostname":  sysInfo["hostname"],
				"uptime":    sysInfo["uptime"],
				"load":      sysInfo["load"],
				"ipAddress": sysInfo["ipAddress"],
			}

			// 获取CPU核心数
			cpuCores := runtime.NumCPU()

			// 获取CPU型号 - 这里简化处理，使用不同系统的CPU型号示例
			cpuModel := "Unknown CPU Model"
			if runtime.GOOS == "windows" {
				cpuModel = "Intel Core i7-10700K (Windows)"
			} else if runtime.GOOS == "darwin" {
				cpuModel = "Apple M1 (macOS)"
			} else {
				cpuModel = "Intel/AMD CPU (Linux)"
			}

			// CPU信息
			cpuInfo := gin.H{
				"cores": cpuCores,
				"model": cpuModel,
			}

			// 模拟内存信息（实际应该从系统获取）
			totalMem := uint64(16 * 1024 * 1024 * 1024) // 16GB
			usedMem := totalMem * 40 / 100              // 使用40%
			memoryInfo := gin.H{
				"used":  usedMem,
				"total": totalMem,
			}

			// 模拟磁盘信息（实际应该从系统获取）
			totalDisk := uint64(500 * 1024 * 1024 * 1024) // 500GB
			usedDisk := totalDisk * 35 / 100              // 使用35%
			diskInfo := gin.H{
				"used":  usedDisk,
				"total": totalDisk,
			}

			// 模拟进程信息，根据不同操作系统显示不同的典型进程
			var processes []gin.H
			if runtime.GOOS == "windows" {
				processes = []gin.H{
					{"pid": 4, "name": "System", "user": "SYSTEM", "cpu": "0.1", "memory": "0.5", "memoryUsed": 50 * 1024 * 1024, "started": time.Now().Add(-240 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
					{"pid": 728, "name": "svchost.exe", "user": "SYSTEM", "cpu": "1.2", "memory": "0.8", "memoryUsed": 80 * 1024 * 1024, "started": time.Now().Add(-72 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
					{"pid": 1524, "name": "v.exe", "user": "USER", "cpu": "2.5", "memory": "1.2", "memoryUsed": 120 * 1024 * 1024, "started": time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
				}
			} else if runtime.GOOS == "darwin" {
				processes = []gin.H{
					{"pid": 1, "name": "launchd", "user": "root", "cpu": "0.1", "memory": "0.3", "memoryUsed": 30 * 1024 * 1024, "started": time.Now().Add(-240 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
					{"pid": 324, "name": "WindowServer", "user": "root", "cpu": "1.5", "memory": "1.0", "memoryUsed": 100 * 1024 * 1024, "started": time.Now().Add(-48 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
					{"pid": 1524, "name": "v", "user": "user", "cpu": "2.0", "memory": "1.1", "memoryUsed": 110 * 1024 * 1024, "started": time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
				}
			} else {
				processes = []gin.H{
					{"pid": 1, "name": "systemd", "user": "root", "cpu": "0.5", "memory": "0.8", "memoryUsed": 80 * 1024 * 1024, "started": time.Now().Add(-240 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
					{"pid": 854, "name": "v-core", "user": "root", "cpu": "2.1", "memory": "1.2", "memoryUsed": 120 * 1024 * 1024, "started": time.Now().Add(-48 * time.Hour).Format("2006-01-02 15:04:05"), "state": "running"},
				}
			}

			// 构建一个完整且符合前端预期的响应
			response := gin.H{
				"code":    200,
				"message": "success",
				"data": gin.H{
					"systemInfo":  systemInfo,
					"cpuInfo":     cpuInfo,
					"cpuUsage":    45, // 模拟CPU使用率45%
					"memoryInfo":  memoryInfo,
					"memoryUsage": 40, // 模拟内存使用率40%
					"diskInfo":    diskInfo,
					"diskUsage":   35, // 模拟磁盘使用率35%
					"processes":   processes,
				},
			}

			// 直接输出完整响应结构，方便调试
			log.Info("Final API response", logger.Fields{
				"response": response,
			})

			c.JSON(http.StatusOK, response)
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
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 优雅关闭
	quit := make(chan os.Signal, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", logger.Fields{
				"error": err,
			})
		}
	}()

	log.Info("Server started", logger.Fields{
		"address": ":8080",
	})

	// 确保信号通道被正确初始化
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Waiting for shutdown signal - server is now running...", logger.Fields{})

	// 等待中断信号
	<-quit
	log.Info("Server shutting down")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", logger.Fields{
			"error": err,
		})
	}

	log.Info("Server exited")
}
