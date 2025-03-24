package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"v/logger"
)

// SiteSettings represents site settings
type SiteSettings struct {
	Name            string `json:"name" env:"SITE_NAME"`
	Description     string `json:"description" env:"SITE_DESCRIPTION"`
	AllowRegister   bool   `json:"allow_register" env:"SITE_ALLOW_REGISTER"`
	MaintenanceMode bool   `json:"maintenance_mode" env:"SITE_MAINTENANCE_MODE"`
}

// TrafficSettings represents traffic settings
type TrafficSettings struct {
	DefaultLimit      int64         `json:"default_limit" env:"TRAFFIC_DEFAULT_LIMIT"`
	StatsInterval     time.Duration `json:"stats_interval" env:"TRAFFIC_STATS_INTERVAL"`
	WarningPercent    int           `json:"warning_percent" env:"TRAFFIC_WARNING_PERCENT"`
	AccountExpireDays int           `json:"account_expire_days" env:"TRAFFIC_ACCOUNT_EXPIRE_DAYS"`
}

// SSLSettings represents SSL settings
type SSLSettings struct {
	AutoRenew         bool          `json:"auto_renew" env:"SSL_AUTO_RENEW"`
	RenewDays         int           `json:"renew_days" env:"SSL_RENEW_DAYS"`
	Provider          string        `json:"provider" env:"SSL_PROVIDER"`
	Email             string        `json:"email" env:"SSL_EMAIL"`
	CertDir           string        `json:"cert_dir" env:"SSL_CERT_DIR"`
	AcmeURL           string        `json:"acme_url" env:"SSL_ACME_URL"`
	ChallengeType     string        `json:"challenge_type" env:"SSL_CHALLENGE_TYPE"`
	CheckInterval     time.Duration `json:"check_interval" env:"SSL_CHECK_INTERVAL"`
	RenewInterval     time.Duration `json:"renew_interval" env:"SSL_RENEW_INTERVAL"`
	ExpiryWarningDays time.Duration `json:"expiry_warning_days" env:"SSL_EXPIRY_WARNING_DAYS"`
	RenewBeforeDays   time.Duration `json:"renew_before_days" env:"SSL_RENEW_BEFORE_DAYS"`
}

// ProxySettings represents proxy settings
type ProxySettings struct {
	DefaultPort    int      `json:"default_port" env:"PROXY_DEFAULT_PORT"`
	AllowedIPs     []string `json:"allowed_ips" env:"PROXY_ALLOWED_IPS"`
	BlockedIPs     []string `json:"blocked_ips" env:"PROXY_BLOCKED_IPS"`
	MaxConnections int      `json:"max_connections" env:"PROXY_MAX_CONNECTIONS"`
}

// SecuritySettings represents security settings
type SecuritySettings struct {
	JWTSecret         string        `json:"jwt_secret" env:"SECURITY_JWT_SECRET"`
	TokenExpiry       time.Duration `json:"token_expiry" env:"SECURITY_TOKEN_EXPIRY"`
	MinPasswordLength int           `json:"min_password_length" env:"SECURITY_MIN_PASSWORD_LENGTH"`
	LoginAttempts     int           `json:"login_attempts" env:"SECURITY_LOGIN_ATTEMPTS"`
	LockoutTime       time.Duration `json:"lockout_time" env:"SECURITY_LOCKOUT_TIME"`
}

// NotificationSettings represents notification settings
type NotificationSettings struct {
	EnableEmail  bool   `json:"enable_email" env:"NOTIFICATION_ENABLE_EMAIL"`
	SMTPHost     string `json:"smtp_host" env:"NOTIFICATION_SMTP_HOST"`
	SMTPPort     int    `json:"smtp_port" env:"NOTIFICATION_SMTP_PORT"`
	SMTPUser     string `json:"smtp_user" env:"NOTIFICATION_SMTP_USER"`
	SMTPPassword string `json:"smtp_password" env:"NOTIFICATION_SMTP_PASSWORD"`
	FromEmail    string `json:"from_email" env:"NOTIFICATION_FROM_EMAIL"`
	FromName     string `json:"from_name" env:"NOTIFICATION_FROM_NAME"`
}

// BackupSettings represents backup settings
type BackupSettings struct {
	Enable      bool          `json:"enable" env:"BACKUP_ENABLE"`
	Interval    time.Duration `json:"interval" env:"BACKUP_INTERVAL"`
	Retention   int           `json:"retention" env:"BACKUP_RETENTION"`
	Path        string        `json:"path" env:"BACKUP_PATH"`
	Compression bool          `json:"compression" env:"BACKUP_COMPRESSION"`
}

// MonitorSettings represents monitor settings
type MonitorSettings struct {
	Interval          time.Duration `json:"interval" env:"MONITOR_INTERVAL"`
	CPUThreshold      float64       `json:"cpu_threshold" env:"MONITOR_CPU_THRESHOLD"`
	MemoryThreshold   float64       `json:"memory_threshold" env:"MONITOR_MEMORY_THRESHOLD"`
	DiskThreshold     float64       `json:"disk_threshold" env:"MONITOR_DISK_THRESHOLD"`
	EnableCPUAlert    bool          `json:"enable_cpu_alert" env:"MONITOR_ENABLE_CPU_ALERT"`
	EnableMemoryAlert bool          `json:"enable_memory_alert" env:"MONITOR_ENABLE_MEMORY_ALERT"`
	EnableDiskAlert   bool          `json:"enable_disk_alert" env:"MONITOR_ENABLE_DISK_ALERT"`
	AlertInterval     int           `json:"alert_interval" env:"MONITOR_ALERT_INTERVAL"`
}

// LogSettings represents log settings
type LogSettings struct {
	Level         string        `json:"level" env:"LOG_LEVEL"`
	ConsoleLog    bool          `json:"console_log" env:"LOG_CONSOLE_LOG"`
	FileLog       bool          `json:"file_log" env:"LOG_FILE_LOG"`
	FilePath      string        `json:"file_path" env:"LOG_FILE_PATH"`
	MaxSize       int           `json:"max_size" env:"LOG_MAX_SIZE"`
	MaxAge        int           `json:"max_age" env:"LOG_MAX_AGE"`
	MaxBackups    int           `json:"max_backups" env:"LOG_MAX_BACKUPS"`
	Compress      bool          `json:"compress" env:"LOG_COMPRESS"`
	ErrorFilePath string        `json:"error_file_path" env:"LOG_ERROR_FILE_PATH"`
	SeparateError bool          `json:"separate_error" env:"LOG_SEPARATE_ERROR"`
	RotateTime    time.Duration `json:"rotate_time" env:"LOG_ROTATE_TIME"`
}

// AdminSettings represents admin settings
type AdminSettings struct {
	Email string `json:"email" env:"ADMIN_EMAIL"`
}

// XraySettings represents xray settings
type XraySettings struct {
	Version       string        `json:"version" env:"XRAY_VERSION"`
	AutoUpdate    bool          `json:"auto_update" env:"XRAY_AUTO_UPDATE"`
	CheckInterval time.Duration `json:"check_interval" env:"XRAY_CHECK_INTERVAL"`
	CustomConfig  bool          `json:"custom_config" env:"XRAY_CUSTOM_CONFIG"`
	ConfigPath    string        `json:"config_path" env:"XRAY_CONFIG_PATH"`
}

// Settings represents system settings
type Settings struct {
	// Site settings
	Site SiteSettings `json:"site"`

	// Traffic settings
	Traffic TrafficSettings `json:"traffic"`

	// SSL settings
	SSL SSLSettings `json:"ssl"`

	// Proxy settings
	Proxy ProxySettings `json:"proxy"`

	// Security settings
	Security SecuritySettings `json:"security"`

	// Notification settings
	Notification NotificationSettings `json:"notification"`

	// Backup settings
	Backup BackupSettings `json:"backup"`

	// Monitor settings
	Monitor MonitorSettings `json:"monitor"`

	// Log settings
	Log LogSettings `json:"log"`

	// Admin settings
	Admin AdminSettings `json:"admin"`

	// Xray settings
	Xray XraySettings `json:"xray"`

	// Protocol settings
	Protocols map[string]bool `json:"protocols"`

	// Transport settings
	Transports map[string]bool `json:"transports"`
}

// Manager represents a settings manager
type Manager struct {
	log          *logger.Logger
	settings     *Settings
	settingsPath string
	mu           sync.RWMutex
}

// New creates a new settings manager
func New(log *logger.Logger) *Manager {
	return &Manager{
		log:          log,
		settings:     &Settings{},
		settingsPath: filepath.Join("config", "settings.json"),
	}
}

// Start starts the settings manager
func (m *Manager) Start() error {
	// Create config directory
	if err := os.MkdirAll(filepath.Dir(m.settingsPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Load settings
	if err := m.Load(); err != nil {
		return fmt.Errorf("failed to load settings: %v", err)
	}

	m.log.Info("Settings manager started", logger.Fields{
		"settings_path": m.settingsPath,
	})

	return nil
}

// Stop stops the settings manager
func (m *Manager) Stop() {
	// Save settings
	if err := m.Save(); err != nil {
		m.log.Error("Failed to save settings", logger.Fields{
			"error": err,
		})
	}
}

// Load loads settings from file and environment variables
func (m *Manager) Load() error {
	// Load from file
	if err := m.loadFromFile(); err != nil {
		m.log.Warn("Failed to load settings from file", logger.Fields{
			"error": err,
		})
	}

	// Load from environment variables
	if err := m.loadFromEnv(); err != nil {
		return fmt.Errorf("failed to load settings from environment: %v", err)
	}

	// 确保协议和传输层设置存在，不为nil
	if m.settings.Protocols == nil {
		m.settings.Protocols = make(map[string]bool)
	}
	if m.settings.Transports == nil {
		m.settings.Transports = make(map[string]bool)
	}

	// 设置默认值
	// 默认协议
	if _, exists := m.settings.Protocols["trojan"]; !exists {
		m.settings.Protocols["trojan"] = true
	}
	if _, exists := m.settings.Protocols["vmess"]; !exists {
		m.settings.Protocols["vmess"] = true
	}
	if _, exists := m.settings.Protocols["vless"]; !exists {
		m.settings.Protocols["vless"] = true
	}
	if _, exists := m.settings.Protocols["shadowsocks"]; !exists {
		m.settings.Protocols["shadowsocks"] = true
	}
	if _, exists := m.settings.Protocols["socks"]; !exists {
		m.settings.Protocols["socks"] = false
	}
	if _, exists := m.settings.Protocols["http"]; !exists {
		m.settings.Protocols["http"] = false
	}

	// 默认传输层
	if _, exists := m.settings.Transports["tcp"]; !exists {
		m.settings.Transports["tcp"] = true
	}
	if _, exists := m.settings.Transports["ws"]; !exists {
		m.settings.Transports["ws"] = true
	}
	if _, exists := m.settings.Transports["http2"]; !exists {
		m.settings.Transports["http2"] = true
	}
	if _, exists := m.settings.Transports["grpc"]; !exists {
		m.settings.Transports["grpc"] = true
	}
	if _, exists := m.settings.Transports["quic"]; !exists {
		m.settings.Transports["quic"] = false
	}

	return nil
}

// loadFromFile loads settings from file
func (m *Manager) loadFromFile() error {
	// Check if file exists
	if _, err := os.Stat(m.settingsPath); os.IsNotExist(err) {
		m.log.Warn("Settings file does not exist, using defaults", logger.Fields{
			"path": m.settingsPath,
		})
		return nil
	}

	// Read file
	data, err := os.ReadFile(m.settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %v", err)
	}

	m.log.Debug("Read settings file content", logger.Fields{
		"content_length": len(data),
		"path":           m.settingsPath,
	})

	// Unmarshal settings
	if err := json.Unmarshal(data, m.settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %v", err)
	}

	// 记录关键设置
	m.log.Info("Settings loaded from file", logger.Fields{
		"auto_update": m.settings.Xray.AutoUpdate,
		"protocols":   m.settings.Protocols != nil,
		"transports":  m.settings.Transports != nil,
	})

	return nil
}

// loadFromEnv loads settings from environment variables
func (m *Manager) loadFromEnv() error {
	val := reflect.ValueOf(m.settings).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip if field is not a struct
		if field.Kind() != reflect.Struct {
			continue
		}

		// Process struct fields
		for j := 0; j < field.NumField(); j++ {
			structField := field.Field(j)
			structType := fieldType.Type.Field(j)

			// Get environment variable name
			envTag := structType.Tag.Get("env")
			if envTag == "" {
				continue
			}

			// Get environment variable value
			envValue := os.Getenv(envTag)
			if envValue == "" {
				continue
			}

			// Set field value based on type
			switch structField.Kind() {
			case reflect.String:
				structField.SetString(envValue)
			case reflect.Int, reflect.Int64:
				if intValue, err := strconv.ParseInt(envValue, 10, 64); err == nil {
					structField.SetInt(intValue)
				}
			case reflect.Float64:
				if floatValue, err := strconv.ParseFloat(envValue, 64); err == nil {
					structField.SetFloat(floatValue)
				}
			case reflect.Bool:
				if boolValue, err := strconv.ParseBool(envValue); err == nil {
					structField.SetBool(boolValue)
				}
			case reflect.Slice:
				if structType.Type.Elem().Kind() == reflect.String {
					structField.Set(reflect.ValueOf(strings.Split(envValue, ",")))
				}
			case reflect.Struct:
				if structType.Type == reflect.TypeOf(time.Duration(0)) {
					if duration, err := time.ParseDuration(envValue); err == nil {
						structField.Set(reflect.ValueOf(duration))
					}
				}
			}
		}
	}

	return nil
}

// Save saves settings to file
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 记录保存前的设置
	m.log.Debug("Saving settings", logger.Fields{
		"auto_update":   m.settings.Xray.AutoUpdate,
		"settings_path": m.settingsPath,
	})

	return m.saveNoLock()
}

// Get returns the current settings
func (m *Manager) Get() *Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy of settings
	settings := *m.settings
	return &settings
}

// Update updates settings
func (m *Manager) Update(settings *Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 记录更新前的重要设置
	oldAutoUpdate := m.settings.Xray.AutoUpdate

	// 记录完整的更新内容
	m.log.Debug("Updating settings", logger.Fields{
		"old_auto_update": m.settings.Xray.AutoUpdate,
		"new_auto_update": settings.Xray.AutoUpdate,
		"xray_settings":   settings.Xray,
	})

	// 手动更新Xray设置
	m.settings.Xray.AutoUpdate = settings.Xray.AutoUpdate
	m.settings.Xray.CheckInterval = settings.Xray.CheckInterval
	m.settings.Xray.CustomConfig = settings.Xray.CustomConfig
	m.settings.Xray.ConfigPath = settings.Xray.ConfigPath
	m.settings.Xray.Version = settings.Xray.Version

	// 手动更新协议和传输层设置
	if settings.Protocols != nil {
		// 如果m.settings.Protocols为nil，先初始化
		if m.settings.Protocols == nil {
			m.settings.Protocols = make(map[string]bool)
		}

		// 复制新的协议设置
		for k, v := range settings.Protocols {
			m.settings.Protocols[k] = v
		}
	}

	if settings.Transports != nil {
		// 如果m.settings.Transports为nil，先初始化
		if m.settings.Transports == nil {
			m.settings.Transports = make(map[string]bool)
		}

		// 复制新的传输层设置
		for k, v := range settings.Transports {
			m.settings.Transports[k] = v
		}
	}

	// 记录变更
	if oldAutoUpdate != m.settings.Xray.AutoUpdate {
		m.log.Info("Xray auto update setting changed", logger.Fields{
			"old_value": oldAutoUpdate,
			"new_value": m.settings.Xray.AutoUpdate,
		})
	}

	// 在解锁前保存设置
	if err := m.saveNoLock(); err != nil {
		return fmt.Errorf("failed to save settings: %v", err)
	}

	// 验证保存后的设置
	fileData, err := os.ReadFile(m.settingsPath)
	if err == nil {
		var savedSettings Settings
		if err := json.Unmarshal(fileData, &savedSettings); err == nil {
			m.log.Debug("Verified saved settings", logger.Fields{
				"auto_update_in_file": savedSettings.Xray.AutoUpdate,
			})
		}
	}

	return nil
}

// saveNoLock saves settings without acquiring the lock
func (m *Manager) saveNoLock() error {
	// 在保存前检查protocols和transports是否有效
	if m.settings.Protocols == nil {
		m.settings.Protocols = make(map[string]bool)
	}
	if m.settings.Transports == nil {
		m.settings.Transports = make(map[string]bool)
	}

	// 添加日志记录关键设置
	m.log.Debug("Settings before saving", logger.Fields{
		"auto_update":      m.settings.Xray.AutoUpdate,
		"check_interval":   m.settings.Xray.CheckInterval,
		"custom_config":    m.settings.Xray.CustomConfig,
		"protocols_count":  len(m.settings.Protocols),
		"transports_count": len(m.settings.Transports),
	})

	// Marshal settings
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %v", err)
	}

	// 验证生成的JSON中是否包含正确的auto_update值
	var testSettings map[string]interface{}
	if err := json.Unmarshal(data, &testSettings); err == nil {
		if xray, ok := testSettings["xray"].(map[string]interface{}); ok {
			if autoUpdate, ok := xray["auto_update"].(bool); ok {
				m.log.Debug("Verified JSON before saving", logger.Fields{
					"auto_update_in_json": autoUpdate,
				})
			}
		}
	}

	// Write file
	if err := os.WriteFile(m.settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	m.log.Info("Settings saved successfully", logger.Fields{
		"settings_path": m.settingsPath,
		"auto_update":   m.settings.Xray.AutoUpdate,
		"size":          len(data),
	})

	return nil
}

// GetString returns a string setting value
func (m *Manager) GetString(path string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val := reflect.ValueOf(m.settings).Elem()
	parts := strings.Split(path, ".")

	for _, part := range parts {
		field := val.FieldByName(part)
		if !field.IsValid() {
			return "", fmt.Errorf("invalid setting path: %s", path)
		}

		if field.Kind() == reflect.Struct {
			val = field
		} else if field.Kind() == reflect.String {
			return field.String(), nil
		} else {
			return "", fmt.Errorf("invalid setting type: %s", path)
		}
	}

	return "", fmt.Errorf("invalid setting path: %s", path)
}

// SetString sets a string setting value
func (m *Manager) SetString(path, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	val := reflect.ValueOf(m.settings).Elem()
	parts := strings.Split(path, ".")

	for _, part := range parts {
		field := val.FieldByName(part)
		if !field.IsValid() {
			return fmt.Errorf("invalid setting path: %s", path)
		}

		if field.Kind() == reflect.Struct {
			val = field
		} else if field.Kind() == reflect.String {
			field.SetString(value)
			return m.Save()
		} else {
			return fmt.Errorf("invalid setting type: %s", path)
		}
	}

	return fmt.Errorf("invalid setting path: %s", path)
}

// GetInt returns an integer setting value
func (m *Manager) GetInt(path string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val := reflect.ValueOf(m.settings).Elem()
	parts := strings.Split(path, ".")

	for _, part := range parts {
		field := val.FieldByName(part)
		if !field.IsValid() {
			return 0, fmt.Errorf("invalid setting path: %s", path)
		}

		if field.Kind() == reflect.Struct {
			val = field
		} else if field.Kind() == reflect.Int || field.Kind() == reflect.Int64 {
			return int(field.Int()), nil
		} else {
			return 0, fmt.Errorf("invalid setting type: %s", path)
		}
	}

	return 0, fmt.Errorf("invalid setting path: %s", path)
}

// SetInt sets an integer setting value
func (m *Manager) SetInt(path string, value int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	val := reflect.ValueOf(m.settings).Elem()
	parts := strings.Split(path, ".")

	for _, part := range parts {
		field := val.FieldByName(part)
		if !field.IsValid() {
			return fmt.Errorf("invalid setting path: %s", path)
		}

		if field.Kind() == reflect.Struct {
			val = field
		} else if field.Kind() == reflect.Int || field.Kind() == reflect.Int64 {
			field.SetInt(int64(value))
			return m.Save()
		} else {
			return fmt.Errorf("invalid setting type: %s", path)
		}
	}

	return fmt.Errorf("invalid setting path: %s", path)
}

// GetBool returns a boolean setting value
func (m *Manager) GetBool(path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val := reflect.ValueOf(m.settings).Elem()
	parts := strings.Split(path, ".")

	for _, part := range parts {
		field := val.FieldByName(part)
		if !field.IsValid() {
			return false, fmt.Errorf("invalid setting path: %s", path)
		}

		if field.Kind() == reflect.Struct {
			val = field
		} else if field.Kind() == reflect.Bool {
			return field.Bool(), nil
		} else {
			return false, fmt.Errorf("invalid setting type: %s", path)
		}
	}

	return false, fmt.Errorf("invalid setting path: %s", path)
}

// SetBool sets a boolean setting value
func (m *Manager) SetBool(path string, value bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	val := reflect.ValueOf(m.settings).Elem()
	parts := strings.Split(path, ".")

	for _, part := range parts {
		field := val.FieldByName(part)
		if !field.IsValid() {
			return fmt.Errorf("invalid setting path: %s", path)
		}

		if field.Kind() == reflect.Struct {
			val = field
		} else if field.Kind() == reflect.Bool {
			field.SetBool(value)
			return m.Save()
		} else {
			return fmt.Errorf("invalid setting type: %s", path)
		}
	}

	return fmt.Errorf("invalid setting path: %s", path)
}

// Backup 备份设置到文件
func (m *Manager) Backup() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 创建备份目录
	backupDir := filepath.Join("backups", "settings")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}

	// 创建带时间戳的备份文件名
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("settings_%s.json", timestamp))

	// 序列化设置
	data, err := json.MarshalIndent(m.settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %v", err)
	}

	// 写入备份文件
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup file: %v", err)
	}

	return backupPath, nil
}

// Restore 从备份文件恢复设置
func (m *Manager) Restore(backupPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// 读取备份文件
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %v", err)
	}

	// 反序列化设置
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %v", err)
	}

	// 更新当前设置
	m.settings = &settings

	// 保存设置
	if err := m.Save(); err != nil {
		return fmt.Errorf("failed to save settings: %v", err)
	}

	return nil
}
