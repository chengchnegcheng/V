package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"v/database"
	"v/logger"

	"github.com/gin-gonic/gin"
)

// 定义全局logger
var log = logger.NewLogger()

// ProxyConfig represents a proxy configuration
type ProxyConfig struct {
	ID        int64                  `json:"id"`
	UserID    int64                  `json:"user_id"`
	Name      string                 `json:"name"`
	Protocol  string                 `json:"protocol"`
	Server    string                 `json:"server"`
	Port      int                    `json:"port"`
	Settings  map[string]interface{} `json:"settings"`
	Enabled   bool                   `json:"enabled"`
	Upload    int64                  `json:"upload"`
	Download  int64                  `json:"download"`
	Remark    string                 `json:"remark"`
	CreatedAt time.Time              `json:"created_at"`
}

// HandleListProxies handles GET /api/proxies
func HandleListProxies(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	var rows *sql.Rows
	var err error
	db := database.GetDB()

	// 获取原生SQL连接
	sqlDB, err := db.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	if isAdmin {
		// Admin can see all proxy configs
		rows, err = sqlDB.Query(`
			SELECT id, user_id, name, protocol, server, port, settings, enabled, upload, download, remark, created_at
			FROM proxies
			ORDER BY id DESC
		`)
	} else {
		// Normal user can only see their own proxy configs
		rows, err = sqlDB.Query(`
			SELECT id, user_id, name, protocol, server, port, settings, enabled, upload, download, remark, created_at
			FROM proxies
			WHERE user_id = ?
			ORDER BY id DESC
		`, userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var proxies []ProxyConfig
	for rows.Next() {
		var proxy ProxyConfig
		var settingsJSON string
		err := rows.Scan(
			&proxy.ID,
			&proxy.UserID,
			&proxy.Name,
			&proxy.Protocol,
			&proxy.Server,
			&proxy.Port,
			&settingsJSON,
			&proxy.Enabled,
			&proxy.Upload,
			&proxy.Download,
			&proxy.Remark,
			&proxy.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Parse settings JSON
		if err := json.Unmarshal([]byte(settingsJSON), &proxy.Settings); err != nil {
			log.Error("Failed to unmarshal proxy settings: %v", err)
			// Continue with empty settings rather than failing
			proxy.Settings = make(map[string]interface{})
		}

		proxies = append(proxies, proxy)
	}

	c.JSON(http.StatusOK, proxies)
}

// HandleCreateProxy handles POST /api/proxy
func HandleCreateProxy(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var proxy ProxyConfig
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the user ID
	proxy.UserID = userID

	// Convert settings to JSON
	settingsJSON, err := json.Marshal(proxy.Settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings format"})
		return
	}

	db := database.GetDB()
	// 获取原生SQL连接
	sqlDB, err := db.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	// Insert into database
	result, err := sqlDB.Exec(`
		INSERT INTO proxies (user_id, name, protocol, server, port, settings, enabled, remark, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, proxy.UserID, proxy.Name, proxy.Protocol, proxy.Server, proxy.Port, string(settingsJSON), proxy.Enabled, proxy.Remark, time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy"})
		return
	}

	id, _ := result.LastInsertId()
	proxy.ID = id

	c.JSON(http.StatusOK, proxy)
}

// HandleGetProxy handles GET /api/proxy/:id
func HandleGetProxy(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var proxy ProxyConfig
	var settingsJSON string
	var query string
	var args []interface{}
	db := database.GetDB()

	// 获取原生SQL连接
	sqlDB, err := db.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	if isAdmin {
		query = `
			SELECT id, user_id, name, protocol, server, port, settings, enabled, upload, download, remark, created_at
			FROM proxies
			WHERE id = ?
		`
		args = []interface{}{proxyID}
	} else {
		query = `
			SELECT id, user_id, name, protocol, server, port, settings, enabled, upload, download, remark, created_at
			FROM proxies
			WHERE id = ? AND user_id = ?
		`
		args = []interface{}{proxyID, userID}
	}

	err = sqlDB.QueryRow(query, args...).Scan(
		&proxy.ID,
		&proxy.UserID,
		&proxy.Name,
		&proxy.Protocol,
		&proxy.Server,
		&proxy.Port,
		&settingsJSON,
		&proxy.Enabled,
		&proxy.Upload,
		&proxy.Download,
		&proxy.Remark,
		&proxy.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Parse settings JSON
	if err := json.Unmarshal([]byte(settingsJSON), &proxy.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid settings format"})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

// HandleUpdateProxy handles PUT /api/proxy/:id
func HandleUpdateProxy(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var proxy ProxyConfig
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the proxy exists and belongs to the user
	var existingProxy struct {
		UserID int64
	}
	var query string
	var args []interface{}
	db := database.GetDB()

	// 获取原生SQL连接
	sqlDB, err := db.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	if isAdmin {
		query = "SELECT user_id FROM proxies WHERE id = ?"
		args = []interface{}{proxyID}
	} else {
		query = "SELECT user_id FROM proxies WHERE id = ? AND user_id = ?"
		args = []interface{}{proxyID, userID}
	}

	err = sqlDB.QueryRow(query, args...).Scan(&existingProxy.UserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Convert settings to JSON
	settingsJSON, err := json.Marshal(proxy.Settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings format"})
		return
	}

	// Update proxy
	result, err := sqlDB.Exec(`
		UPDATE proxies
		SET name = ?, protocol = ?, server = ?, port = ?, settings = ?, enabled = ?, remark = ?
		WHERE id = ?
	`, proxy.Name, proxy.Protocol, proxy.Server, proxy.Port, string(settingsJSON), proxy.Enabled, proxy.Remark, proxyID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proxy"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy updated successfully"})
}

// HandleDeleteProxy handles DELETE /api/proxy/:id
func HandleDeleteProxy(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	// Check if the proxy exists and belongs to the user
	var query string
	var args []interface{}
	db := database.GetDB()

	// 获取原生SQL连接
	sqlDB, err := db.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	if isAdmin {
		query = "SELECT 1 FROM proxies WHERE id = ?"
		args = []interface{}{proxyID}
	} else {
		query = "SELECT 1 FROM proxies WHERE id = ? AND user_id = ?"
		args = []interface{}{proxyID, userID}
	}

	var exists bool
	err = sqlDB.QueryRow(query, args...).Scan(&exists)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Delete the proxy
	_, err = sqlDB.Exec("DELETE FROM proxies WHERE id = ?", proxyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete proxy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy deleted successfully"})
}

// HandleGetProxyLink handles GET /api/proxy/:id/link
func HandleGetProxyLink(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	proxyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	// Get proxy details
	var proxy ProxyConfig
	var settingsJSON string
	var query string
	var args []interface{}
	db := database.GetDB()

	// 获取原生SQL连接
	sqlDB, err := db.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	if isAdmin {
		query = `
			SELECT id, user_id, name, protocol, server, port, settings, enabled, remark
			FROM proxies
			WHERE id = ?
		`
		args = []interface{}{proxyID}
	} else {
		query = `
			SELECT id, user_id, name, protocol, server, port, settings, enabled, remark
			FROM proxies
			WHERE id = ? AND user_id = ?
		`
		args = []interface{}{proxyID, userID}
	}

	err = sqlDB.QueryRow(query, args...).Scan(
		&proxy.ID,
		&proxy.UserID,
		&proxy.Name,
		&proxy.Protocol,
		&proxy.Server,
		&proxy.Port,
		&settingsJSON,
		&proxy.Enabled,
		&proxy.Remark,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Parse settings JSON
	if err := json.Unmarshal([]byte(settingsJSON), &proxy.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid settings format"})
		return
	}

	// Generate sharing link based on protocol
	var link string
	var err2 error

	switch strings.ToLower(proxy.Protocol) {
	case "shadowsocks", "ss":
		link, err2 = generateShadowsocksLink(proxy)
	case "trojan":
		link, err2 = generateTrojanLink(proxy)
	case "vmess":
		link, err2 = generateVmessLink(proxy)
	case "vless":
		link, err2 = generateVlessLink(proxy)
	default:
		err2 = fmt.Errorf("unsupported protocol: %s", proxy.Protocol)
	}

	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"link": link})
}

// Helper functions to generate sharing links

func generateShadowsocksLink(proxy ProxyConfig) (string, error) {
	password, ok := proxy.Settings["password"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid password")
	}

	method, ok := proxy.Settings["method"].(string)
	if !ok {
		method = "aes-256-gcm" // default method
	}

	// Base64 encode the user info
	userInfo := fmt.Sprintf("%s:%s", method, password)
	encodedUserInfo := base64.StdEncoding.EncodeToString([]byte(userInfo))

	// Create the ss:// link
	link := fmt.Sprintf("ss://%s@%s:%d", encodedUserInfo, proxy.Server, proxy.Port)

	// Add the name as a hash parameter if it exists
	if proxy.Name != "" {
		encodedName := base64.StdEncoding.EncodeToString([]byte(proxy.Name))
		link = fmt.Sprintf("%s#%s", link, encodedName)
	}

	return link, nil
}

func generateTrojanLink(proxy ProxyConfig) (string, error) {
	password, ok := proxy.Settings["password"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid password")
	}

	sni, _ := proxy.Settings["sni"].(string)

	// Create the trojan:// link with URL-escaped password
	link := fmt.Sprintf("trojan://%s@%s:%d", url.QueryEscape(password), proxy.Server, proxy.Port)

	// Add parameters if they exist
	params := make([]string, 0)

	// Always include security=tls for Trojan
	params = append(params, "security=tls")

	if sni != "" {
		params = append(params, fmt.Sprintf("sni=%s", sni))
	}

	// Add allowInsecure parameter if it exists
	if allowInsecure, ok := proxy.Settings["allowInsecure"].(bool); ok {
		params = append(params, fmt.Sprintf("allowInsecure=%t", allowInsecure))
	}

	if len(params) > 0 {
		link = fmt.Sprintf("%s?%s", link, strings.Join(params, "&"))
	}

	// Add the name as a hash parameter if it exists
	if proxy.Name != "" {
		link = fmt.Sprintf("%s#%s", link, url.QueryEscape(proxy.Name))
	}

	return link, nil
}

func generateVmessLink(proxy ProxyConfig) (string, error) {
	uuid, ok := proxy.Settings["uuid"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid uuid")
	}

	// Get additional settings with defaults
	security, _ := proxy.Settings["security"].(string)
	if security == "" {
		security = "auto"
	}

	network, _ := proxy.Settings["network"].(string)
	if network == "" {
		network = "tcp"
	}

	// Build VMess config object
	config := map[string]interface{}{
		"v":    "2",
		"ps":   proxy.Name,
		"add":  proxy.Server,
		"port": proxy.Port,
		"id":   uuid,
		"aid":  0,
		"net":  network,
		"type": "none",
		"host": "",
		"path": "",
		"tls":  "",
		"scy":  security,
	}

	// Add additional network-specific settings
	switch network {
	case "ws":
		if path, ok := proxy.Settings["path"].(string); ok {
			config["path"] = path
		}
		if host, ok := proxy.Settings["host"].(string); ok {
			config["host"] = host
		}
	case "grpc":
		if serviceName, ok := proxy.Settings["serviceName"].(string); ok {
			config["path"] = serviceName
		}
	}

	// Add TLS settings if enabled
	if tls, ok := proxy.Settings["tls"].(bool); ok && tls {
		config["tls"] = "tls"
		if sni, ok := proxy.Settings["sni"].(string); ok {
			config["sni"] = sni
		}
	}

	// Convert to JSON and encode
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	encodedConfig := base64.StdEncoding.EncodeToString(configJSON)
	return fmt.Sprintf("vmess://%s", encodedConfig), nil
}

func generateVlessLink(proxy ProxyConfig) (string, error) {
	uuid, ok := proxy.Settings["uuid"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid uuid")
	}

	// Create the vless:// link
	link := fmt.Sprintf("vless://%s@%s:%d", uuid, proxy.Server, proxy.Port)

	// Build parameters
	params := []string{"type=tcp"}

	// Add encryption (usually "none" for VLESS)
	params = append(params, "encryption=none")

	// Add TLS settings if enabled
	if tls, ok := proxy.Settings["tls"].(bool); ok && tls {
		params = append(params, "security=tls")
		if sni, ok := proxy.Settings["sni"].(string); ok && sni != "" {
			params = append(params, fmt.Sprintf("sni=%s", sni))
		}
	}

	// Add network-specific settings
	if network, ok := proxy.Settings["network"].(string); ok && network != "" {
		params = append(params, fmt.Sprintf("type=%s", network))

		switch network {
		case "ws":
			if path, ok := proxy.Settings["path"].(string); ok && path != "" {
				params = append(params, fmt.Sprintf("path=%s", path))
			}
			if host, ok := proxy.Settings["host"].(string); ok && host != "" {
				params = append(params, fmt.Sprintf("host=%s", host))
			}
		case "grpc":
			if serviceName, ok := proxy.Settings["serviceName"].(string); ok && serviceName != "" {
				params = append(params, fmt.Sprintf("serviceName=%s", serviceName))
			}
		}
	}

	// Add parameters to the link
	link = fmt.Sprintf("%s?%s", link, strings.Join(params, "&"))

	// Add the name as a hash parameter if it exists
	if proxy.Name != "" {
		link = fmt.Sprintf("%s#%s", link, proxy.Name)
	}

	return link, nil
}
