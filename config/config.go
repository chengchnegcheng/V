package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	Server struct {
		Host     string `json:"host" env:"SERVER_HOST"`
		Port     int    `json:"port" env:"SERVER_PORT"`
		BaseURL  string `json:"base_url" env:"SERVER_BASE_URL"`
		LogLevel string `json:"log_level" env:"SERVER_LOG_LEVEL"`
	} `json:"server"`

	// Database configuration
	Database struct {
		Type     string `json:"type" env:"DB_TYPE"`
		Host     string `json:"host" env:"DB_HOST"`
		Port     int    `json:"port" env:"DB_PORT"`
		User     string `json:"user" env:"DB_USER"`
		Password string `json:"password" env:"DB_PASSWORD"`
		Name     string `json:"name" env:"DB_NAME"`
		SSLMode  string `json:"ssl_mode" env:"DB_SSL_MODE"`
	} `json:"database"`

	// SSL configuration
	SSL struct {
		AutoRenew bool   `json:"auto_renew" env:"SSL_AUTO_RENEW"`
		Email     string `json:"email" env:"SSL_EMAIL"`
		CertDir   string `json:"cert_dir" env:"SSL_CERT_DIR"`
	} `json:"ssl"`

	// Proxy configuration
	Proxy struct {
		DefaultPort int    `json:"default_port" env:"PROXY_DEFAULT_PORT"`
		AllowedIPs  string `json:"allowed_ips" env:"PROXY_ALLOWED_IPS"`
	} `json:"proxy"`

	// Security configuration
	Security struct {
		JWTSecret     string `json:"jwt_secret" env:"JWT_SECRET"`
		TokenExpiry   int    `json:"token_expiry" env:"TOKEN_EXPIRY"`
		AllowRegister bool   `json:"allow_register" env:"ALLOW_REGISTER"`
	} `json:"security"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// Load from file if exists
	if err := loadFromFile(configPath, config); err != nil {
		return nil, fmt.Errorf("failed to load config file: %v", err)
	}

	// Load from environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load config from env: %v", err)
	}

	return config, nil
}

// loadFromFile loads configuration from a JSON file
func loadFromFile(path string, config *Config) error {
	if path == "" {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) error {
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip if not a struct
		if field.Kind() != reflect.Struct {
			continue
		}

		// Process nested struct
		for j := 0; j < field.NumField(); j++ {
			nestedField := field.Field(j)
			nestedType := fieldType.Type.Field(j)

			// Get env tag
			envTag := nestedType.Tag.Get("env")
			if envTag == "" {
				continue
			}

			// Get environment variable value
			envValue := os.Getenv(envTag)
			if envValue == "" {
				continue
			}

			// Set field value based on type
			switch nestedField.Kind() {
			case reflect.String:
				nestedField.SetString(envValue)
			case reflect.Int:
				var intVal int64
				if _, err := fmt.Sscanf(envValue, "%d", &intVal); err != nil {
					return fmt.Errorf("invalid integer value for %s: %v", envTag, err)
				}
				nestedField.SetInt(intVal)
			case reflect.Bool:
				boolVal := strings.ToLower(envValue) == "true"
				nestedField.SetBool(boolVal)
			default:
				return fmt.Errorf("unsupported field type for %s: %v", envTag, nestedField.Kind())
			}
		}
	}

	return nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
