package server

import (
	"v/server/handlers"
	"v/server/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 添加中间件
	r.Use(middleware.Cors())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())

	// 健康检查
	r.GET("/health", handlers.HandleHealth)

	// 用户认证路由
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", handlers.HandleLogin)
		authGroup.POST("/register", handlers.HandleRegister)
		authGroup.POST("/logout", handlers.HandleLogout)
	}

	// 用户管理路由
	userGroup := r.Group("/api/users")
	userGroup.Use(middleware.AuthRequired())
	{
		userGroup.GET("/me", handlers.HandleGetCurrentUser)
		userGroup.PUT("/me", handlers.HandleUpdateCurrentUser)
		userGroup.PUT("/me/password", handlers.HandleUpdatePassword)
	}

	// 协议管理路由
	protocolGroup := r.Group("/api/protocols")
	protocolGroup.Use(middleware.AuthRequired())
	{
		protocolGroup.POST("", handlers.HandleCreateProtocol)
		protocolGroup.GET("/:id", handlers.HandleGetProtocol)
		protocolGroup.GET("", handlers.HandleListProtocols)
		protocolGroup.PUT("/:id", handlers.HandleUpdateProtocol)
		protocolGroup.DELETE("/:id", handlers.HandleDeleteProtocol)
		protocolGroup.POST("/:id/enable", handlers.HandleEnableProtocol)
		protocolGroup.POST("/:id/disable", handlers.HandleDisableProtocol)
	}

	// 证书管理路由
	certGroup := r.Group("/api/certificates")
	certGroup.Use(middleware.AuthRequired())
	{
		certGroup.POST("", handlers.HandleCreateCertificate)
		certGroup.GET("/:id", handlers.HandleGetCertificate)
		certGroup.GET("", handlers.HandleListCertificates)
		certGroup.DELETE("/:id", handlers.HandleDeleteCertificate)
		certGroup.POST("/:id/renew", handlers.HandleRenewCertificate)
		certGroup.POST("/:id/validate", handlers.HandleValidateCertificate)
	}

	// 日志管理路由
	logGroup := r.Group("/api/logs")
	logGroup.Use(middleware.AuthRequired())
	{
		logGroup.GET("", handlers.HandleListLogs)
		logGroup.GET("/export", handlers.HandleExportLogs)
		logGroup.DELETE("", handlers.HandleDeleteLogs)
	}

	// 备份管理路由
	backupGroup := r.Group("/api/backups")
	backupGroup.Use(middleware.AuthRequired())
	{
		backupGroup.POST("", handlers.HandleCreateBackup)
		backupGroup.GET("", handlers.HandleListBackups)
		backupGroup.POST("/:id/restore", handlers.HandleRestoreBackup)
		backupGroup.DELETE("/:id", handlers.HandleDeleteBackup)
		backupGroup.GET("/:id/download", handlers.HandleDownloadBackup)
	}

	// 系统监控路由
	monitorGroup := r.Group("/api/monitor")
	monitorGroup.Use(middleware.AuthRequired())
	{
		monitorGroup.GET("/stats", handlers.HandleGetSystemStats)
	}

	// 流量统计路由
	stats := r.Group("/api/stats")
	stats.Use(middleware.AuthRequired())
	{
		stats.GET("/users/:id", handlers.HandleGetUserStats)
		stats.GET("/protocols/:id", handlers.HandleGetProtocolStats)
		stats.POST("/protocols/:id/traffic", handlers.HandleUpdateProtocolTraffic)
	}

	return r
}
