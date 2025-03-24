package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"v/db"
	"v/errors"
	"v/logger"
	"v/middleware"
	"v/settings"
	"v/xray"
)

// Handler represents an API handler
type Handler struct {
	log        *logger.Logger
	router     *mux.Router
	handlers   map[string]http.HandlerFunc
	db         *db.DB
	settings   *settings.Manager
	xrayMgr    *xray.Manager
	httpServer *http.Server
}

// New creates a new API handler
func New(log *logger.Logger, db *db.DB, settingsMgr *settings.Manager, xrayMgr *xray.Manager) *Handler {
	return &Handler{
		log:      log,
		router:   mux.NewRouter(),
		handlers: make(map[string]http.HandlerFunc),
		db:       db,
		settings: settingsMgr,
		xrayMgr:  xrayMgr,
	}
}

// Start starts the API server
func (h *Handler) Start() error {
	// Setup routes
	h.Setup()

	// Setup xray version endpoints
	h.setupXrayEndpoints()

	// Setup proxy sharing endpoints
	h.setupShareEndpoints()

	// Setup protocol settings endpoints
	h.setupProtocolEndpoints()

	// Setup inbound management endpoints
	h.setupInboundEndpoints()

	// Start HTTP server
	h.httpServer = &http.Server{
		Addr:    ":9000",
		Handler: h.router,
	}

	go func() {
		if err := h.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.log.Error("API server error", logger.Fields{
				"error": err,
			})
		}
	}()

	h.log.Info("API server started", logger.Fields{
		"address": ":9000",
	})

	return nil
}

// Stop stops the API server
func (h *Handler) Stop() error {
	if h.httpServer != nil {
		return h.httpServer.Close()
	}
	return nil
}

// Register registers a new handler
func (h *Handler) Register(path string, handler http.HandlerFunc) {
	h.handlers[path] = handler
}

// Setup sets up the API routes
func (h *Handler) Setup() {
	// Add middleware
	h.router.Use(middleware.ToMuxMiddleware(middleware.Logging(h.log)))
	h.router.Use(middleware.ToMuxMiddleware(middleware.Recovery(h.log)))
	h.router.Use(middleware.ToMuxMiddleware(middleware.CORS()))
	h.router.Use(middleware.ToMuxMiddleware(middleware.RateLimit()))

	// Register handlers
	for path, handler := range h.handlers {
		h.router.HandleFunc(path, handler)
	}

	// Add not found handler
	h.router.NotFoundHandler = http.HandlerFunc(h.handleNotFound)
}

// setupXrayEndpoints sets up the xray version management endpoints
func (h *Handler) setupXrayEndpoints() {
	// Get supported versions
	h.router.HandleFunc("/api/xray/versions", func(w http.ResponseWriter, r *http.Request) {
		currentVersion := h.xrayMgr.GetCurrentVersion()
		supportedVersions := h.xrayMgr.GetSupportedVersions()

		// 确保至少有一个版本可用
		if len(supportedVersions) == 0 {
			supportedVersions = []string{"v1.8.24", "v1.8.23", "v1.8.22"}
			h.log.Warn("No supported versions available, using default list", logger.Fields{})
		}

		h.handleResponse(w, map[string]interface{}{
			"current_version":    currentVersion,
			"supported_versions": supportedVersions,
		})
	}).Methods("GET")

	// Switch version
	h.router.HandleFunc("/api/xray/version", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Version string `json:"version"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		if req.Version == "" {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		if err := h.xrayMgr.SwitchVersion(req.Version); err != nil {
			h.handleError(w, err)
			return
		}

		h.handleResponse(w, map[string]interface{}{
			"success":         true,
			"current_version": req.Version,
		})
	}).Methods("POST")

	// Start Xray
	h.router.HandleFunc("/api/xray/start", func(w http.ResponseWriter, r *http.Request) {
		if err := h.xrayMgr.Start(); err != nil {
			h.handleError(w, err)
			return
		}

		h.handleResponse(w, map[string]interface{}{
			"success": true,
		})
	}).Methods("POST")

	// Stop Xray
	h.router.HandleFunc("/api/xray/stop", func(w http.ResponseWriter, r *http.Request) {
		if err := h.xrayMgr.Stop(); err != nil {
			h.handleError(w, err)
			return
		}

		h.handleResponse(w, map[string]interface{}{
			"success": true,
		})
	}).Methods("POST")

	// Get Xray status
	h.router.HandleFunc("/api/xray/status", func(w http.ResponseWriter, r *http.Request) {
		h.handleResponse(w, map[string]interface{}{
			"running":         h.xrayMgr.IsRunning(),
			"current_version": h.xrayMgr.GetCurrentVersion(),
		})
	}).Methods("GET")

	// Update Xray settings
	h.router.HandleFunc("/api/settings/xray", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AutoUpdate    bool   `json:"auto_update"`
			CustomConfig  bool   `json:"custom_config"`
			ConfigPath    string `json:"config_path"`
			CheckInterval int    `json:"check_interval"`
		}

		if r.Method == "GET" {
			// Get current settings
			settings := h.settings.Get()
			h.handleResponse(w, map[string]interface{}{
				"auto_update":    settings.Xray.AutoUpdate,
				"custom_config":  settings.Xray.CustomConfig,
				"config_path":    settings.Xray.ConfigPath,
				"check_interval": settings.Xray.CheckInterval / time.Hour,
			})
			return
		}

		// POST - Update settings
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		// 获取当前设置的完整拷贝
		settings := h.settings.Get()

		// 更新Xray相关设置
		settings.Xray.AutoUpdate = req.AutoUpdate
		settings.Xray.CustomConfig = req.CustomConfig
		if req.ConfigPath != "" {
			settings.Xray.ConfigPath = req.ConfigPath
		}
		if req.CheckInterval > 0 {
			settings.Xray.CheckInterval = time.Duration(req.CheckInterval) * time.Hour
		}

		// 使用Update方法更新并保存所有设置
		if err := h.settings.Update(settings); err != nil {
			h.handleError(w, err)
			return
		}

		h.handleResponse(w, map[string]interface{}{
			"success": true,
		})
	}).Methods("GET", "POST")

	// Test custom config
	h.router.HandleFunc("/api/xray/test-config", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ConfigPath string `json:"config_path"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		if req.ConfigPath == "" {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		// Check if file exists and is valid JSON
		if _, err := os.Stat(req.ConfigPath); os.IsNotExist(err) {
			h.handleError(w, errors.ErrResourceNotFound)
			return
		}

		// Read file
		data, err := os.ReadFile(req.ConfigPath)
		if err != nil {
			h.handleError(w, errors.ErrInternalServerError)
			return
		}

		// Validate JSON
		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		h.handleResponse(w, map[string]interface{}{
			"success": true,
		})
	}).Methods("POST")
}

// setupShareEndpoints sets up the proxy sharing endpoints
func (h *Handler) setupShareEndpoints() {
	// Get proxy share link
	h.router.HandleFunc("/api/proxy/{id}/link", func(w http.ResponseWriter, r *http.Request) {
		id := h.getPathParam(r, "id")
		if id == "" {
			h.handleError(w, errors.ErrMissingParameter)
			return
		}

		// 获取代理信息
		// 实际项目中应该从数据库查询
		// 这里为演示使用模拟数据
		var proxyInfo map[string]interface{}

		switch id {
		case "1", "2":
			proxyInfo = map[string]interface{}{
				"id":            id,
				"type":          "trojan",
				"name":          "trojan" + id,
				"host":          "example.com",
				"port":          443,
				"password":      "password123",
				"sni":           "example.com",
				"allowInsecure": false,
			}
		case "3", "4":
			var port int
			if id == "3" {
				port = 60606
			} else {
				port = 60605
			}
			proxyInfo = map[string]interface{}{
				"id":            id,
				"type":          "trojan",
				"name":          "trojan",
				"host":          "example.com",
				"port":          port,
				"password":      "password123",
				"sni":           "example.com",
				"allowInsecure": false,
			}
		default:
			h.handleError(w, errors.ErrResourceNotFound)
			return
		}

		// 根据不同协议类型生成不同的分享链接
		var shareLink string
		switch proxyInfo["type"] {
		case "trojan":
			password := proxyInfo["password"].(string)
			host := proxyInfo["host"].(string)
			port := proxyInfo["port"]
			sni := proxyInfo["sni"].(string)
			name := proxyInfo["name"].(string)

			// 构建Trojan分享链接
			// trojan://password@host:port?sni=sni#name
			shareLink = fmt.Sprintf("trojan://%s@%s:%v?sni=%s&allowInsecure=%v#%s",
				password, host, port, sni,
				proxyInfo["allowInsecure"],
				url.PathEscape(name))

		case "vmess":
			// 这里添加VMess链接生成逻辑
			shareLink = "vmess://示例VMess链接"

		case "vless":
			// 这里添加VLESS链接生成逻辑
			shareLink = "vless://示例VLESS链接"

		case "shadowsocks":
			// 这里添加Shadowsocks链接生成逻辑
			shareLink = "ss://示例Shadowsocks链接"

		default:
			h.handleError(w, errors.ErrInvalidParameter)
			return
		}

		h.handleResponse(w, map[string]interface{}{
			"link": shareLink,
		})
	}).Methods("GET")

	// Get proxy QR code
	h.router.HandleFunc("/api/proxy/{id}/qrcode", func(w http.ResponseWriter, r *http.Request) {
		id := h.getPathParam(r, "id")
		if id == "" {
			h.handleError(w, errors.ErrMissingParameter)
			return
		}

		// 先获取分享链接
		var link string

		// 实际项目中应该查询数据库
		// 这里简化处理，直接构造链接
		switch id {
		case "1", "2":
			link = fmt.Sprintf("trojan://password123@example.com:443?sni=example.com&allowInsecure=false#trojan%s", id)
		case "3":
			link = "trojan://password123@example.com:60606?sni=example.com&allowInsecure=false#trojan"
		case "4":
			link = "trojan://password123@example.com:60605?sni=example.com&allowInsecure=false#trojan"
		default:
			h.handleError(w, errors.ErrResourceNotFound)
			return
		}

		// 设置QR码图像大小
		size := 256
		if sizeParam := h.getQueryParam(r, "size"); sizeParam != "" {
			if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 1024 {
				size = s
			}
		}

		// 生成QR码
		// 这里返回链接的Base64编码，前端可以使用这个显示QR码
		// 实际项目中可以直接返回图像数据
		h.handleResponse(w, map[string]interface{}{
			"qrcode": link, // 实际项目中应该生成真正的QR码图像
			"link":   link,
			"size":   size,
		})
	}).Methods("GET")
}

// setupProtocolEndpoints 设置协议管理相关API
func (h *Handler) setupProtocolEndpoints() {
	// 获取协议设置
	h.router.HandleFunc("/api/settings/protocols", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// 获取当前设置
			settings := h.settings.Get()

			// 默认协议设置
			protocols := map[string]bool{
				"trojan":      true,
				"vmess":       true,
				"vless":       true,
				"shadowsocks": true,
				"socks":       false,
				"http":        false,
			}

			// 默认传输层设置
			transports := map[string]bool{
				"tcp":   true,
				"ws":    true,
				"http2": true,
				"grpc":  true,
				"quic":  false,
			}

			// 如果有保存的设置，使用保存的设置
			if settings.Protocols != nil {
				for k, v := range settings.Protocols {
					protocols[k] = v
				}
			}

			if settings.Transports != nil {
				for k, v := range settings.Transports {
					transports[k] = v
				}
			}

			h.handleResponse(w, map[string]interface{}{
				"protocols":  protocols,
				"transports": transports,
			})
			return
		}

		// POST - 更新设置
		var req struct {
			Protocols  map[string]bool `json:"protocols"`
			Transports map[string]bool `json:"transports"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		// 更新设置
		settings := h.settings.Get()
		settings.Protocols = req.Protocols
		settings.Transports = req.Transports

		// 保存设置
		if err := h.settings.Update(settings); err != nil {
			h.handleError(w, err)
			return
		}

		// 更新Xray配置（如有必要）
		if h.xrayMgr.IsRunning() {
			h.log.Info("Protocols or transports changed, Xray config needs to be updated", logger.Fields{
				"protocols":  req.Protocols,
				"transports": req.Transports,
			})

			// 这里可以添加更新Xray配置的逻辑
			// 通常需要根据启用的协议和传输层生成新的配置
		}

		h.handleResponse(w, map[string]interface{}{
			"success": true,
		})
	}).Methods("GET", "POST")
}

// setupInboundEndpoints 设置入站管理相关API
func (h *Handler) setupInboundEndpoints() {
	// 获取所有入站
	h.router.HandleFunc("/api/inbounds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// 获取所有入站列表
			// 这里应该从数据库获取，示例使用模拟数据
			inbounds := []map[string]interface{}{
				{
					"id":           1,
					"remark":       "测试节点1",
					"protocol":     "shadowsocks",
					"port":         10086,
					"listen":       "0.0.0.0",
					"enable":       true,
					"network":      "tcp+udp",
					"traffic_up":   1024 * 1024 * 10,   // 10MB
					"traffic_down": 1024 * 1024 * 1024, // 1GB
					"created_at":   time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
					"updated_at":   time.Now().Format(time.RFC3339),
				},
				{
					"id":           2,
					"remark":       "测试节点2",
					"protocol":     "trojan",
					"port":         443,
					"listen":       "0.0.0.0",
					"enable":       true,
					"network":      "tcp",
					"traffic_up":   1024 * 1024 * 100,  // 100MB
					"traffic_down": 1024 * 1024 * 2048, // 2GB
					"created_at":   time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
					"updated_at":   time.Now().Format(time.RFC3339),
				},
				{
					"id":           3,
					"remark":       "测试节点3",
					"protocol":     "vmess",
					"port":         8080,
					"listen":       "0.0.0.0",
					"enable":       false,
					"network":      "tcp",
					"traffic_up":   1024 * 1024 * 50,  // 50MB
					"traffic_down": 1024 * 1024 * 500, // 500MB
					"created_at":   time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
					"updated_at":   time.Now().Format(time.RFC3339),
				},
			}

			h.handleResponse(w, inbounds)
			return
		}

		// POST - 添加入站
		var inbound map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&inbound); err != nil {
			h.handleError(w, errors.ErrInvalidRequestBody)
			return
		}

		// 验证必填字段
		required := []string{"remark", "protocol", "port"}
		for _, field := range required {
			if _, ok := inbound[field]; !ok {
				h.handleError(w, errors.ErrMissingParameter)
				return
			}
		}

		// 这里应该添加到数据库，示例直接返回成功
		h.log.Info("添加入站", logger.Fields{
			"inbound": inbound,
		})

		h.handleResponse(w, map[string]interface{}{
			"success": true,
			"message": "入站添加成功",
			"id":      time.Now().Unix(), // 模拟生成ID
		})
	}).Methods("GET", "POST")

	// 获取单个入站
	h.router.HandleFunc("/api/inbounds/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := h.getPathParam(r, "id")
		if id == "" {
			h.handleError(w, errors.ErrMissingParameter)
			return
		}

		// 根据ID获取入站信息
		// 这里应该从数据库查询，示例使用模拟数据
		var inbound map[string]interface{}

		switch id {
		case "1":
			inbound = map[string]interface{}{
				"id":           1,
				"remark":       "测试节点1",
				"protocol":     "shadowsocks",
				"port":         10086,
				"listen":       "0.0.0.0",
				"enable":       true,
				"network":      "tcp+udp",
				"traffic_up":   1024 * 1024 * 10,   // 10MB
				"traffic_down": 1024 * 1024 * 1024, // 1GB
				"created_at":   time.Now().AddDate(0, 0, -5).Format(time.RFC3339),
				"updated_at":   time.Now().Format(time.RFC3339),
			}
		case "2":
			inbound = map[string]interface{}{
				"id":           2,
				"remark":       "测试节点2",
				"protocol":     "trojan",
				"port":         443,
				"listen":       "0.0.0.0",
				"enable":       true,
				"network":      "tcp",
				"traffic_up":   1024 * 1024 * 100,  // 100MB
				"traffic_down": 1024 * 1024 * 2048, // 2GB
				"created_at":   time.Now().AddDate(0, 0, -10).Format(time.RFC3339),
				"updated_at":   time.Now().Format(time.RFC3339),
			}
		case "3":
			inbound = map[string]interface{}{
				"id":           3,
				"remark":       "测试节点3",
				"protocol":     "vmess",
				"port":         8080,
				"listen":       "0.0.0.0",
				"enable":       false,
				"network":      "tcp",
				"traffic_up":   1024 * 1024 * 50,  // 50MB
				"traffic_down": 1024 * 1024 * 500, // 500MB
				"created_at":   time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
				"updated_at":   time.Now().Format(time.RFC3339),
			}
		default:
			h.handleError(w, errors.ErrResourceNotFound)
			return
		}

		if r.Method == "GET" {
			h.handleResponse(w, inbound)
		} else if r.Method == "PUT" {
			// 更新入站
			var updateData map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
				h.handleError(w, errors.ErrInvalidRequestBody)
				return
			}

			// 这里应该更新数据库
			h.log.Info("更新入站", logger.Fields{
				"id":         id,
				"updateData": updateData,
			})

			h.handleResponse(w, map[string]interface{}{
				"success": true,
				"message": "入站更新成功",
			})
		} else if r.Method == "DELETE" {
			// 删除入站
			// 这里应该从数据库删除
			h.log.Info("删除入站", logger.Fields{
				"id": id,
			})

			h.handleResponse(w, map[string]interface{}{
				"success": true,
				"message": "入站删除成功",
			})
		}
	}).Methods("GET", "PUT", "DELETE")

	// 获取入站链接
	h.router.HandleFunc("/api/inbounds/{id}/link", func(w http.ResponseWriter, r *http.Request) {
		id := h.getPathParam(r, "id")
		if id == "" {
			h.handleError(w, errors.ErrMissingParameter)
			return
		}

		// 根据ID和协议生成链接
		// 这里应该从数据库查询入站信息，示例使用模拟数据
		var link string

		switch id {
		case "1":
			// shadowsocks链接
			link = "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQxMjM=@example.com:10086#%E6%B5%8B%E8%AF%95%E8%8A%82%E7%82%B91"
		case "2":
			// trojan链接
			link = "trojan://password123@example.com:443?security=tls&sni=example.com#%E6%B5%8B%E8%AF%95%E8%8A%82%E7%82%B92"
		case "3":
			// vmess链接
			link = "vmess://eyJhZGQiOiJleGFtcGxlLmNvbSIsImFpZCI6IjAiLCJob3N0IjoiIiwiaWQiOiI4M2M2NGJlMS0xZTQ3LTRhNmEtOTkyYi1iODI1ZGVjYTVjNmMiLCJuZXQiOiJ0Y3AiLCJwYXRoIjoiIiwicG9ydCI6IjgwODAiLCJwcyI6Iua1i%2be6p%2BiKgueCuTMiLCJzY3kiOiJhdXRvIiwic25pIjoiIiwidGxzIjoiIiwidHlwZSI6Im5vbmUiLCJ2IjoiMiJ9"
		default:
			h.handleError(w, errors.ErrResourceNotFound)
			return
		}

		h.handleResponse(w, map[string]interface{}{
			"link": link,
		})
	}).Methods("GET")
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// handleNotFound handles the not found route
func (h *Handler) handleNotFound(w http.ResponseWriter, r *http.Request) {
	h.handleError(w, errors.ErrResourceNotFound)
}

// handleError handles the error response
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	// Log the error
	h.log.Error("API error", logger.Fields{
		"error": err.Error(),
	})

	// Handle custom errors
	if e, ok := err.(*errors.Error); ok {
		w.WriteHeader(e.Code)
		h.handleResponse(w, map[string]interface{}{
			"error": e.Message,
		})
		return
	}

	// Handle standard errors
	w.WriteHeader(http.StatusInternalServerError)
	h.handleResponse(w, map[string]interface{}{
		"error": "Internal server error",
	})
}

// handleResponse handles API responses
func (h *Handler) handleResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// getPathParam gets a path parameter
func (h *Handler) getPathParam(r *http.Request, name string) string {
	return mux.Vars(r)[name]
}

// getQueryParam gets a query parameter
func (h *Handler) getQueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// getAuthToken gets the authentication token
func (h *Handler) getAuthToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// getContentType gets the content type
func (h *Handler) getContentType(r *http.Request) string {
	return r.Header.Get("Content-Type")
}

// getUserAgent gets the user agent
func (h *Handler) getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// getIP gets the client IP
func (h *Handler) getIP(r *http.Request) string {
	// Try X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Try X-Forwarded-For header
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}

	// Use remote address
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
