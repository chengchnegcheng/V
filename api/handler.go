package api

import (
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"v/cert"
	"v/monitor"
	"v/protocol"
	"v/router"
	"v/traffic"
)

// APIHandler API处理器
type APIHandler struct {
	logger   *slog.Logger
	db       interface{}
	traffic  traffic.Manager
	protocol protocol.Manager
	cert     cert.CertificateManager
	monitor  monitor.SystemMonitor
}

// NewHandler 创建新的API处理器
func NewHandler(
	logger *slog.Logger,
	db interface{},
	traffic traffic.Manager,
	protocol protocol.Manager,
	cert cert.CertificateManager,
	monitor monitor.SystemMonitor,
) *APIHandler {
	return &APIHandler{
		logger:   logger,
		db:       db,
		traffic:  traffic,
		protocol: protocol,
		cert:     cert,
		monitor:  monitor,
	}
}

// RegisterRoutes 注册API路由
func (h *APIHandler) RegisterRoutes(r router.Router) {
	api := r.Group("/api")

	// 健康检查
	api.GET("/health", h.handleHealth)

	// 系统相关
	api.GET("/system/info", h.handleSystemInfo)
}

// handleHealth 处理健康检查
func (h *APIHandler) handleHealth(c *router.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// handleSystemInfo 处理系统信息
func (h *APIHandler) handleSystemInfo(c *router.Context) {
	// 获取系统信息
	info := map[string]interface{}{
		"status": "running",
		"os":     runtime.GOOS,
		"arch":   runtime.GOARCH,
	}

	// 添加更多详细信息
	hostname, err := os.Hostname()
	if err == nil {
		info["hostname"] = hostname
	}

	info["go_version"] = runtime.Version()
	info["num_cpu"] = runtime.NumCPU()
	info["num_goroutine"] = runtime.NumGoroutine()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info["memory_alloc"] = m.Alloc / 1024 / 1024            // MB
	info["memory_total_alloc"] = m.TotalAlloc / 1024 / 1024 // MB
	info["memory_sys"] = m.Sys / 1024 / 1024                // MB

	// 返回信息
	c.JSON(http.StatusOK, info)
}
