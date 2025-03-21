package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"v/api"
	"v/cert"
	"v/logger"
	"v/model"
	"v/monitor"
	"v/notification"
	"v/protocol"
	"v/settings"
	"v/traffic"

	"log/slog"

	"github.com/gin-gonic/gin"
)

var (
	configFile    string
	webRoot       string
	listenAddress string
)

func init() {
	flag.StringVar(&configFile, "config", "config/settings.json", "配置文件路径")
	flag.StringVar(&webRoot, "webroot", "web", "Web根目录")
	flag.StringVar(&listenAddress, "listen", ":8080", "监听地址")
	flag.Parse()
}

func main() {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		fmt.Printf("创建日志目录失败: %v\n", err)
		os.Exit(1)
	}

	// 创建日志器
	log := logger.NewLogger()
	defer log.Close()

	// 创建日志分析器
	analyzer := logger.NewLogAnalyzer("logs")

	// 创建设置管理器
	settingsMgr := settings.New(log)
	if err := settingsMgr.Start(); err != nil {
		log.Fatal("启动设置管理器失败: %v", err)
	}
	defer settingsMgr.Stop()

	// 获取设置
	// 注释掉未使用的变量
	// config := settingsMgr.Get()

	// 创建数据库连接
	db, err := initDatabase()
	if err != nil {
		log.Fatal("初始化数据库失败: %v", err)
	}

	// 创建通知器
	notificationConfig := settingsMgr.Get().Notification
	notifier := notification.NewEmailNotifier(
		notificationConfig.SMTPHost,
		notificationConfig.SMTPPort,
		notificationConfig.SMTPUser,
		notificationConfig.SMTPPassword,
		notificationConfig.FromEmail,
		notificationConfig.FromName,
	)

	// 创建日志管理器
	logMgr := logger.NewManager(log, db)
	if err := logMgr.Start(); err != nil {
		log.Fatal("启动日志管理器失败: %v", err)
	}
	defer logMgr.Stop()

	// 创建监控管理器
	monitorMgr := monitor.New(log, settingsMgr, notifier, db)
	if err := monitorMgr.Start(); err != nil {
		log.Fatal("启动监控管理器失败: %v", err)
	}
	defer monitorMgr.Stop()

	// 创建流量管理器
	stdLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	trafficMgr := traffic.New(stdLogger, db, notifier)
	trafficMgr.Start()
	defer trafficMgr.Stop()

	// 创建协议管理器
	protocolMgr := protocol.New(log, settingsMgr, db)

	// 创建证书管理器
	// 使用新的 CertManager 类型名称
	certMgr := cert.NewCertManager(log, settingsMgr, notifier, db, webRoot)
	if err := certMgr.Start(); err != nil {
		log.Fatal("启动证书管理器失败: %v", err)
	}
	defer certMgr.Stop()

	// 创建API服务器
	router := gin.Default()

	// 注册API处理器
	apiGroup := router.Group("/api")

	// 系统信息API
	sysHandler := api.NewSystemHandler(log, monitorMgr)
	sysHandler.RegisterRoutes(apiGroup)

	// 证书API
	certHandler := api.NewCertificateHandler(certMgr)
	certHandler.RegisterRoutes(apiGroup)

	// 日志API
	logHandler := api.NewLogHandler(log, db, logMgr, analyzer)
	logHandler.RegisterRoutes(apiGroup)

	// 流量API
	trafficHandler := api.NewTrafficHandler(log, trafficMgr)
	trafficHandler.RegisterRoutes(apiGroup)

	// 协议API
	protocolHandler := api.NewProtocolHandler(log, protocolMgr)
	protocolHandler.RegisterRoutes(apiGroup)

	// 设置API
	settingsHandler := api.NewSettingsHandler(log, settingsMgr)
	settingsHandler.RegisterRoutes(apiGroup)

	// 启动HTTP服务器
	go func() {
		log.Info("启动HTTP服务器在 %s", listenAddress)
		if err := router.Run(listenAddress); err != nil {
			log.Fatal("启动HTTP服务器失败: %v", err)
		}
	}()

	// 等待信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info("收到退出信号，正在关闭服务...")

	// 给服务一些时间进行清理
	time.Sleep(time.Second)
}

// initDatabase 初始化数据库
func initDatabase() (model.DB, error) {
	// 创建数据目录
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %v", err)
	}

	dbPath := filepath.Join("data", "v.db")
	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 初始化数据库
	sqliteDB := model.NewSQLiteDB(db, slog.Default())

	// 自动迁移
	if err := sqliteDB.AutoMigrate(); err != nil {
		return nil, err
	}

	return sqliteDB, nil
}
