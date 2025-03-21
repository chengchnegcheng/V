package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"v/database"
	"v/proxy"

	"github.com/gin-gonic/gin"
)

// ProxyProtocol represents supported proxy protocols
type ProxyProtocol string

const (
	ProtocolVMess       ProxyProtocol = "vmess"
	ProtocolVLESS       ProxyProtocol = "vless"
	ProtocolTrojan      ProxyProtocol = "trojan"
	ProtocolShadowsocks ProxyProtocol = "shadowsocks"
)

// ProxyRequest represents the request body for creating/updating a proxy
type ProxyRequest struct {
	Protocol ProxyProtocol   `json:"protocol" binding:"required,oneof=vmess vless trojan shadowsocks"`
	Settings json.RawMessage `json:"settings" binding:"required"`
}

type ProxyConfig struct {
	ID        int64                  `json:"id"`
	UserID    int64                  `json:"user_id"`
	Protocol  string                 `json:"protocol"`
	Settings  map[string]interface{} `json:"settings"`
	Enabled   bool                   `json:"enabled"`
	Upload    int64                  `json:"upload"`
	Download  int64                  `json:"download"`
	CreatedAt string                 `json:"created_at"`
}

// HandleListProxyConfigs handles GET /api/proxy
func HandleListProxyConfigs(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	var rows *sql.Rows
	var err error

	if isAdmin {
		// Admin can see all proxy configs
		rows, err = database.DB.Query(`
			SELECT id, user_id, protocol, settings, enabled, upload, download, created_at
			FROM proxy_configs
			ORDER BY id DESC
		`)
	} else {
		// Normal user can only see their own proxy configs
		rows, err = database.DB.Query(`
			SELECT id, user_id, protocol, settings, enabled, upload, download, created_at
			FROM proxy_configs
			WHERE user_id = ?
			ORDER BY id DESC
		`, userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var configs []ProxyConfig
	for rows.Next() {
		var config ProxyConfig
		var settingsJSON string
		err := rows.Scan(
			&config.ID,
			&config.UserID,
			&config.Protocol,
			&settingsJSON,
			&config.Enabled,
			&config.Upload,
			&config.Download,
			&config.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Parse settings JSON
		if err := json.Unmarshal([]byte(settingsJSON), &config.Settings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid settings format"})
			return
		}

		configs = append(configs, config)
	}

	c.JSON(http.StatusOK, configs)
}

// HandleCreateProxyConfig handles POST /api/proxy
func HandleCreateProxyConfig(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var config ProxyConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate protocol
	switch config.Protocol {
	case "vmess", "vless", "trojan", "shadowsocks":
		// Valid protocol
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid protocol"})
		return
	}

	// Convert settings to JSON
	settingsJSON, err := json.Marshal(config.Settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings format"})
		return
	}

	// Insert proxy config
	result, err := database.DB.Exec(`
		INSERT INTO proxy_configs (user_id, protocol, settings, enabled, created_at)
		VALUES (?, ?, ?, 1, CURRENT_TIMESTAMP)
	`, userID, config.Protocol, string(settingsJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy config"})
		return
	}

	configID, _ := result.LastInsertId()

	// Start proxy server
	if err := proxy.DefaultService.StartProxy(configID, userID, config.Protocol, config.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start proxy server: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Proxy config created successfully",
		"id":      configID,
	})
}

// HandleGetProxyConfig handles GET /api/proxy/:id
func HandleGetProxyConfig(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var config ProxyConfig
	var settingsJSON string
	var query string
	var args []interface{}

	if isAdmin {
		query = `
			SELECT id, user_id, protocol, settings, enabled, upload, download, created_at
			FROM proxy_configs
			WHERE id = ?
		`
		args = []interface{}{configID}
	} else {
		query = `
			SELECT id, user_id, protocol, settings, enabled, upload, download, created_at
			FROM proxy_configs
			WHERE id = ? AND user_id = ?
		`
		args = []interface{}{configID, userID}
	}

	err = database.DB.QueryRow(query, args...).Scan(
		&config.ID,
		&config.UserID,
		&config.Protocol,
		&settingsJSON,
		&config.Enabled,
		&config.Upload,
		&config.Download,
		&config.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy config not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Parse settings JSON
	if err := json.Unmarshal([]byte(settingsJSON), &config.Settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid settings format"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// HandleUpdateProxyConfig handles PUT /api/proxy/:id
func HandleUpdateProxyConfig(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var config ProxyConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the proxy config exists and belongs to the user
	var existingConfig ProxyConfig
	var query string
	var args []interface{}

	if isAdmin {
		query = "SELECT user_id FROM proxy_configs WHERE id = ?"
		args = []interface{}{configID}
	} else {
		query = "SELECT user_id FROM proxy_configs WHERE id = ? AND user_id = ?"
		args = []interface{}{configID, userID}
	}

	err = database.DB.QueryRow(query, args...).Scan(&existingConfig.UserID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy config not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Convert settings to JSON
	settingsJSON, err := json.Marshal(config.Settings)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings format"})
		return
	}

	// Update proxy config
	result, err := database.DB.Exec(`
		UPDATE proxy_configs
		SET protocol = ?, settings = ?, enabled = ?
		WHERE id = ?
	`, config.Protocol, string(settingsJSON), config.Enabled, configID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update proxy config"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy config not found"})
		return
	}

	// Restart proxy server if enabled
	if config.Enabled {
		if err := proxy.DefaultService.StopProxy(configID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop proxy server: " + err.Error()})
			return
		}
		if err := proxy.DefaultService.StartProxy(configID, existingConfig.UserID, config.Protocol, config.Settings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start proxy server: " + err.Error()})
			return
		}
	} else {
		if err := proxy.DefaultService.StopProxy(configID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop proxy server: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy config updated successfully"})
}

// HandleDeleteProxyConfig handles DELETE /api/proxy/:id
func HandleDeleteProxyConfig(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	// Stop proxy server
	if err := proxy.DefaultService.StopProxy(configID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop proxy server: " + err.Error()})
		return
	}

	// Delete proxy config
	var query string
	var args []interface{}

	if isAdmin {
		query = "DELETE FROM proxy_configs WHERE id = ?"
		args = []interface{}{configID}
	} else {
		query = "DELETE FROM proxy_configs WHERE id = ? AND user_id = ?"
		args = []interface{}{configID, userID}
	}

	result, err := database.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete proxy config"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proxy config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proxy config deleted successfully"})
}

// validateProxySettings validates the settings for a specific protocol
func validateProxySettings(protocol ProxyProtocol, settings json.RawMessage) bool {
	var settingsMap map[string]interface{}
	if err := json.Unmarshal(settings, &settingsMap); err != nil {
		return false
	}

	switch protocol {
	case ProtocolVMess:
		return validateVMessSettings(settingsMap)
	case ProtocolVLESS:
		return validateVLESSSettings(settingsMap)
	case ProtocolTrojan:
		return validateTrojanSettings(settingsMap)
	case ProtocolShadowsocks:
		return validateShadowsocksSettings(settingsMap)
	default:
		return false
	}
}

func validateVMessSettings(settings map[string]interface{}) bool {
	required := []string{"id", "alterId", "security"}
	for _, field := range required {
		if _, ok := settings[field]; !ok {
			return false
		}
	}
	return true
}

func validateVLESSSettings(settings map[string]interface{}) bool {
	required := []string{"id", "encryption"}
	for _, field := range required {
		if _, ok := settings[field]; !ok {
			return false
		}
	}
	return true
}

func validateTrojanSettings(settings map[string]interface{}) bool {
	required := []string{"password"}
	for _, field := range required {
		if _, ok := settings[field]; !ok {
			return false
		}
	}
	return true
}

func validateShadowsocksSettings(settings map[string]interface{}) bool {
	required := []string{"method", "password"}
	for _, field := range required {
		if _, ok := settings[field]; !ok {
			return false
		}
	}
	return true
}
