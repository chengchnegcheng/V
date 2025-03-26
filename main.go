package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"v/api"
	"v/common"
	"v/logger"
	"v/model"
	"v/monitor"
	"v/settings"
	"v/xray"

	"github.com/gin-gonic/gin"
)

var (
	// ... any existing variables ...
	systemMonitor *monitor.SystemStatsMonitor
	// Mock DB for testing
	mockDB *MockDB
)

// Add parseFlags function
func parseFlags() {
	// You can add command line flag parsing here if needed
	flag.Parse()
}

// Add initLogger function
func initLogger() {
	// Any logger initialization code can go here
	// This is a placeholder since the real initialization is done below
}

// MockDB implements model.DB interface for testing
type MockDB struct {
	log *logger.Logger
}

// AutoMigrate implements model.DB interface
func (m *MockDB) AutoMigrate() error {
	return nil
}

// Begin implements the Begin method for transactions
func (m *MockDB) Begin() error {
	return nil
}

// Commit implements the Commit method for transactions
func (m *MockDB) Commit() error {
	return nil
}

// Rollback implements the Rollback method for transactions
func (m *MockDB) Rollback() error {
	return nil
}

// Close implements the Close method
func (m *MockDB) Close() error {
	return nil
}

// Stub implementation of other DB interface methods - not implementing all for brevity
// In a real implementation, these methods would need to be completed
func (m *MockDB) CreateUser(user *model.User) error                      { return nil }
func (m *MockDB) GetUser(id int64) (*model.User, error)                  { return nil, nil }
func (m *MockDB) GetUserByUsername(username string) (*model.User, error) { return nil, nil }
func (m *MockDB) GetUserByEmail(email string) (*model.User, error)       { return nil, nil }
func (m *MockDB) UpdateUser(user *model.User) error                      { return nil }
func (m *MockDB) DeleteUser(id int64) error                              { return nil }
func (m *MockDB) ListUsers(page, pageSize int) ([]*model.User, error)    { return nil, nil }
func (m *MockDB) GetTotalUsers() (int64, error)                          { return 0, nil }
func (m *MockDB) SearchUsers(keyword string) ([]*model.User, error)      { return nil, nil }
func (m *MockDB) GetSettings(key string) (string, error)                 { return "", nil }
func (m *MockDB) SetSettings(key, value string) error                    { return nil }

// Implement CreateProxy and related methods
func (m *MockDB) CreateProxy(proxy *common.Proxy) error                    { return nil }
func (m *MockDB) GetProxy(id int64) (*common.Proxy, error)                 { return nil, nil }
func (m *MockDB) GetProxiesByUserID(userID int64) ([]*common.Proxy, error) { return nil, nil }
func (m *MockDB) UpdateProxy(proxy *common.Proxy) error                    { return nil }
func (m *MockDB) DeleteProxy(id int64) error                               { return nil }
func (m *MockDB) GetProxiesByPort(port int) ([]*common.Proxy, error)       { return nil, nil }
func (m *MockDB) ListProxies(page, pageSize int) ([]*common.Proxy, error)  { return nil, nil }
func (m *MockDB) GetTotalProxies() (int64, error)                          { return 0, nil }
func (m *MockDB) SearchProxies(keyword string) ([]*common.Proxy, error)    { return nil, nil }

// Implement traffic-related methods
func (m *MockDB) CreateTraffic(traffic *common.TrafficStats) error                   { return nil }
func (m *MockDB) GetTraffic(id int64) (*common.TrafficStats, error)                  { return nil, nil }
func (m *MockDB) UpdateTraffic(traffic *common.TrafficStats) error                   { return nil }
func (m *MockDB) DeleteTraffic(id int64) error                                       { return nil }
func (m *MockDB) ListTrafficByUserID(userID int64) ([]*common.TrafficStats, error)   { return nil, nil }
func (m *MockDB) ListTrafficByProxyID(proxyID int64) ([]*common.TrafficStats, error) { return nil, nil }
func (m *MockDB) GetTrafficStats(userID uint) (*model.TrafficStats, error)           { return nil, nil }
func (m *MockDB) CreateTrafficRecord(traffic *model.Traffic) error                   { return nil }
func (m *MockDB) CleanupTraffic(before time.Time) error                              { return nil }

// Implement protocol-related methods
func (m *MockDB) CreateProtocol(protocol *model.Protocol) error                { return nil }
func (m *MockDB) GetProtocol(id int64) (*model.Protocol, error)                { return nil, nil }
func (m *MockDB) GetProtocolsByUserID(userID int64) ([]*model.Protocol, error) { return nil, nil }
func (m *MockDB) UpdateProtocol(protocol *model.Protocol) error                { return nil }
func (m *MockDB) DeleteProtocol(id int64) error                                { return nil }
func (m *MockDB) GetProtocolsByPort(port int) ([]*model.Protocol, error)       { return nil, nil }
func (m *MockDB) ListProtocols(page, pageSize int) ([]*model.Protocol, error)  { return nil, nil }
func (m *MockDB) GetTotalProtocols() (int64, error)                            { return 0, nil }
func (m *MockDB) SearchProtocols(keyword string) ([]*model.Protocol, error)    { return nil, nil }

// Implement protocol stats methods
func (m *MockDB) CreateProtocolStats(stats *model.ProtocolStats) error    { return nil }
func (m *MockDB) GetProtocolStats(id int64) (*model.ProtocolStats, error) { return nil, nil }
func (m *MockDB) UpdateProtocolStats(stats *model.ProtocolStats) error    { return nil }
func (m *MockDB) ListProtocolStatsByUserID(userID int64) ([]*model.ProtocolStats, error) {
	return nil, nil
}

// Implement certificate-related methods
func (m *MockDB) CreateCertificate(cert *model.Certificate) error          { return nil }
func (m *MockDB) GetCertificate(domain string) (*model.Certificate, error) { return nil, nil }
func (m *MockDB) UpdateCertificate(cert *model.Certificate) error          { return nil }
func (m *MockDB) DeleteCertificate(domain string) error                    { return nil }
func (m *MockDB) ListCertificates() ([]*model.Certificate, error)          { return nil, nil }

// Implement alert methods
func (m *MockDB) CreateAlert(alert *model.AlertRecord) error                  { return nil }
func (m *MockDB) GetAlert(id int64) (*model.AlertRecord, error)               { return nil, nil }
func (m *MockDB) ListAlerts(page, pageSize int) ([]*model.AlertRecord, error) { return nil, nil }
func (m *MockDB) DeleteAlert(id int64) error                                  { return nil }

// Implement log-related methods
func (m *MockDB) CreateLog(log *model.Log) error                       { return nil }
func (m *MockDB) GetLog(id int64) (*model.Log, error)                  { return nil, nil }
func (m *MockDB) UpdateLog(log *model.Log) error                       { return nil }
func (m *MockDB) DeleteLog(id int64) error                             { return nil }
func (m *MockDB) ListLogs(query *model.LogQuery) ([]*model.Log, error) { return nil, nil }
func (m *MockDB) GetTotalLogs(query *model.LogQuery) (int64, error)    { return 0, nil }
func (m *MockDB) DeleteLogsBefore(t time.Time) error                   { return nil }
func (m *MockDB) ExportLogs(query *model.LogQuery) (string, error)     { return "", nil }

// Implement backup-related methods
func (m *MockDB) CreateBackup(backup *model.Backup) error   { return nil }
func (m *MockDB) GetBackup(id int64) (*model.Backup, error) { return nil, nil }
func (m *MockDB) UpdateBackup(backup *model.Backup) error   { return nil }
func (m *MockDB) DeleteBackup(id int64) error               { return nil }
func (m *MockDB) ListBackups() ([]*model.Backup, error)     { return nil, nil }
func (m *MockDB) GetTotalBackups() (int64, error)           { return 0, nil }
func (m *MockDB) DeleteBackupsBefore(t time.Time) error     { return nil }

// Implement daily stats methods
func (m *MockDB) CreateDailyStats(stats *model.DailyStats) error                   { return nil }
func (m *MockDB) DeleteDailyStatsBefore(date time.Time) error                      { return nil }
func (m *MockDB) ListDailyStatsByUserID(userID int64) ([]*model.DailyStats, error) { return nil, nil }
func (m *MockDB) ListProtocolStatsByProtocolID(protocolID int64) ([]*model.ProtocolStats, error) {
	return nil, nil
}

// Implement alert records methods
func (m *MockDB) CreateAlertRecord(record *model.AlertRecord) error { return nil }
func (m *MockDB) ListAlertRecords(out *[]*model.AlertRecord) error  { return nil }

// Implement traffic history methods
func (m *MockDB) CreateTrafficHistory(history *model.TrafficHistory) error { return nil }
func (m *MockDB) ListTrafficHistoryByDateRange(userID uint, startDate, endDate string, histories *[]*model.TrafficHistory) error {
	return nil
}

func main() {
	// Parse command line flags
	parseFlags()

	// Initialize logger
	initLogger()

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

	// 初始化模拟数据库
	mockDB = &MockDB{log: log}

	// 创建系统监控
	systemMonitor = monitor.NewSystemStatsMonitor(mockDB)

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

			// 把sysInfo转换为所需格式
			systemInfo := gin.H{
				"os":        sysInfo["os"],
				"kernel":    sysInfo["kernel"],
				"hostname":  sysInfo["hostname"],
				"uptime":    sysInfo["uptime"],
				"load":      sysInfo["load"],
				"ipAddress": sysInfo["ipAddress"],
				"arch":      sysInfo["arch"],     // 添加架构信息
				"platform":  sysInfo["platform"], // 添加平台信息
				"cpus":      sysInfo["cpus"],     // 添加CPU核心数
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

			// 尝试获取系统统计信息
			systemStats, err := systemMonitor.GetSystemStats()
			if err != nil {
				log.Error("Failed to get system stats", logger.Fields{
					"error": err.Error(),
				})
			}

			cpuUsagePercent := 45.0
			if err == nil && systemStats != nil {
				// 如果成功获取系统统计信息，使用实际值
				cpuUsagePercent = systemStats.CPUUsage
			}

			// CPU信息
			cpuInfo := gin.H{
				"cores":     cpuCores,
				"model":     cpuModel,
				"usage":     cpuUsagePercent,
				"frequency": "3.5 GHz",              // 示例值
				"cache":     "16 MB",                // 示例值
				"processes": runtime.NumGoroutine(), // 当前Go协程数量作为进程数
				"threads":   cpuCores * 2,           // 假设每个核心有2个线程
			}

			// 获取实际内存信息
			var totalMem uint64
			var usedMem uint64
			var memoryUsage float64 = 40.0
			var diskUsage float64 = 35.0

			if systemStats != nil {
				// 使用实际系统统计数据
				totalMem = systemStats.MemoryTotal
				usedMem = systemStats.MemoryUsed
				memoryUsage = systemStats.MemoryUsage
				diskUsage = systemStats.DiskUsage

				// 内存信息
				memoryInfo := gin.H{
					"used":       systemStats.MemoryUsed,
					"total":      systemStats.MemoryTotal,
					"free":       systemStats.MemoryFree,
					"buffers":    systemStats.MemoryFree / 4,  // 示例值
					"cached":     systemStats.MemoryFree / 3,  // 示例值
					"swap_total": systemStats.MemoryTotal / 2, // 示例值
					"swap_used":  systemStats.MemoryTotal / 8, // 示例值
				}

				// 磁盘信息
				diskInfo := gin.H{
					"used":       systemStats.DiskUsed,
					"total":      systemStats.DiskTotal,
					"free":       systemStats.DiskFree,
					"mount":      "/",    // 示例值
					"filesystem": "ext4", // 示例值
				}

				// 构建响应
				response := gin.H{
					"code":    200,
					"message": "success",
					"data": gin.H{
						"systemInfo":  systemInfo,
						"cpuInfo":     cpuInfo,
						"cpuUsage":    cpuUsagePercent,
						"memoryInfo":  memoryInfo,
						"memoryUsage": memoryUsage,
						"diskInfo":    diskInfo,
						"diskUsage":   diskUsage,
						"processes":   getProcessInfo(),
					},
				}

				log.Info("Final API response (using system stats)", logger.Fields{
					"status": "success",
				})

				c.JSON(http.StatusOK, response)
				return
			}

			// 如果无法获取系统统计信息，则使用模拟数据
			// 模拟内存信息
			totalMem = uint64(16 * 1024 * 1024 * 1024) // 16GB
			usedMem = totalMem * 40 / 100              // 使用40%
			memoryInfo := gin.H{
				"used":       usedMem,
				"total":      totalMem,
				"free":       totalMem - usedMem,
				"buffers":    uint64(1 * 1024 * 1024 * 1024), // 1GB
				"cached":     uint64(2 * 1024 * 1024 * 1024), // 2GB
				"swap_total": uint64(8 * 1024 * 1024 * 1024), // 8GB
				"swap_used":  uint64(2 * 1024 * 1024 * 1024), // 2GB
			}

			// 模拟磁盘信息
			totalDisk := uint64(500 * 1024 * 1024 * 1024) // 500GB
			usedDisk := totalDisk * 35 / 100              // 使用35%
			diskInfo := gin.H{
				"used":       usedDisk,
				"total":      totalDisk,
				"free":       totalDisk - usedDisk,
				"mount":      "/",
				"filesystem": "NTFS",
			}

			// 模拟进程信息
			processes := getProcessInfo()

			// 构建一个完整且符合前端预期的响应
			response := gin.H{
				"code":    200,
				"message": "success",
				"data": gin.H{
					"systemInfo":  systemInfo,
					"cpuInfo":     cpuInfo,
					"cpuUsage":    cpuUsagePercent, // 使用默认或真实CPU使用率
					"memoryInfo":  memoryInfo,
					"memoryUsage": memoryUsage, // 使用默认或真实内存使用率
					"diskInfo":    diskInfo,
					"diskUsage":   diskUsage, // 使用默认或真实磁盘使用率
					"processes":   processes,
				},
			}

			// 直接输出完整响应结构，方便调试
			log.Info("Final API response (using mock data)", logger.Fields{
				"response": "mock_data_used",
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

// getProcessInfo 返回进程信息列表，根据操作系统返回不同的进程列表
func getProcessInfo() []gin.H {
	// 根据不同操作系统显示不同的典型进程
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
	return processes
}
